// Package util provides utility functions for the bundle inspector.
package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

	// Open source file
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
