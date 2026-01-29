// Package util provides utility functions for the bundle inspector.
package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/compression"
)

// ExtractZip extracts a ZIP archive to a temporary directory and returns the path.
func ExtractZip(zipPath string) (string, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "bundle-inspector-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Open ZIP file
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer r.Close()

	// Extract files
	for _, f := range r.File {
		if err := extractZipFile(f, tempDir); err != nil {
			os.RemoveAll(tempDir)
			return "", err
		}
	}

	return tempDir, nil
}

func extractZipFile(f *zip.File, destDir string) error {
	// Construct target path
	targetPath := filepath.Join(destDir, f.Name)

	// Check for zip slip vulnerability
	if !filepath.HasPrefix(targetPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", f.Name)
	}

	// Create directory if needed
	if f.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, os.ModePerm)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return err
	}

	// Create file
	outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Check if file uses LZFSE compression (method 99)
	if f.Method == compression.CompressionMethodLZFSE {
		// Read compressed data using OpenRaw (returns io.Reader)
		rawReader, err := f.OpenRaw()
		if err != nil {
			return fmt.Errorf("failed to open LZFSE compressed file: %w", err)
		}

		compressedData, err := io.ReadAll(rawReader)
		if err != nil {
			return fmt.Errorf("failed to read LZFSE compressed data: %w", err)
		}

		// Decompress with LZFSE
		decompressed, err := compression.DecompressLZFSE(compressedData)
		if err != nil {
			return fmt.Errorf("LZFSE decompression failed for %s: %w\nHint: This IPA uses LZFSE compression (method 99). Make sure LZFSE support is enabled.", f.Name, err)
		}

		// Write decompressed content
		_, err = io.Copy(outFile, bytes.NewReader(decompressed))
		return err
	}

	// Standard decompression (DEFLATE, STORE, etc.)
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Copy contents
	_, err = io.Copy(outFile, rc)
	return err
}

// WalkDirectory walks a directory tree and calls the callback for each file.
func WalkDirectory(root string, callback func(path string, info os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return callback(path, info)
		}
		return nil
	})
}
