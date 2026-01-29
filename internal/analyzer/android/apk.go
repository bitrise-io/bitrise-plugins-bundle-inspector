// Package android provides analyzers for Android artifacts.
package android

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// APKAnalyzer analyzes Android APK files.
type APKAnalyzer struct{}

// NewAPKAnalyzer creates a new APK analyzer.
func NewAPKAnalyzer() *APKAnalyzer {
	return &APKAnalyzer{}
}

// ValidateArtifact checks if the file is a valid APK.
func (a *APKAnalyzer) ValidateArtifact(path string) error {
	if !strings.HasSuffix(strings.ToLower(path), ".apk") {
		return fmt.Errorf("file must have .apk extension")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	return nil
}

// Analyze performs analysis on an APK file.
func (a *APKAnalyzer) Analyze(ctx context.Context, path string) (*types.Report, error) {
	if err := a.ValidateArtifact(path); err != nil {
		return nil, err
	}

	// Get APK file size
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat APK: %w", err)
	}

	// Open APK as ZIP
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open APK: %w", err)
	}
	defer zipReader.Close()

	// Parse manifest for metadata
	manifest, err := parseManifest(path)
	if err != nil {
		// Non-fatal, continue without manifest data
		manifest = make(map[string]interface{})
	}

	// Build file tree and calculate sizes
	fileTree, uncompressedSize := util.BuildZipFileTree(&zipReader.Reader)

	// Create size breakdown
	sizeBreakdown := categorizeAPKSizes(fileTree)

	// Find largest files
	largestFiles := util.FindLargestFiles(fileTree, 10)

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:             path,
			Type:             types.ArtifactTypeAPK,
			Size:             info.Size(),
			UncompressedSize: uncompressedSize,
			AnalyzedAt:       time.Now(),
		},
		SizeBreakdown: sizeBreakdown,
		FileTree:      fileTree,
		LargestFiles:  largestFiles,
		Metadata:      manifest,
	}

	return report, nil
}

// parseManifest extracts basic information from AndroidManifest.xml
func parseManifest(apkPath string) (map[string]interface{}, error) {
	manifest := make(map[string]interface{})

	// Open APK file
	zipFile, err := zip.OpenReader(apkPath)
	if err != nil {
		return manifest, err
	}
	defer zipFile.Close()

	// Find AndroidManifest.xml
	for _, f := range zipFile.File {
		if f.Name == "AndroidManifest.xml" {
			manifest["has_manifest"] = true
			break
		}
	}

	// Note: Full manifest parsing requires binary XML decoding
	// For MVP, we just detect its presence
	// Full parsing can be added in future phases

	return manifest, nil
}

// categorizeAPKSizes creates a size breakdown for APK
func categorizeAPKSizes(nodes []*types.FileNode) types.SizeBreakdown {
	breakdown := types.SizeBreakdown{
		ByCategory:  make(map[string]int64),
		ByExtension: make(map[string]int64),
	}

	var categorizeNode func(node *types.FileNode)
	categorizeNode = func(node *types.FileNode) {
		if node.IsDir {
			dirName := strings.ToLower(node.Name)

			// Categorize by directory
			if dirName == "lib" {
				breakdown.Libraries += node.Size
				breakdown.ByCategory["Native Libraries"] += node.Size
			} else if dirName == "res" {
				breakdown.Resources += node.Size
				breakdown.ByCategory["Resources"] += node.Size
			} else if dirName == "assets" {
				breakdown.Assets += node.Size
				breakdown.ByCategory["Assets"] += node.Size
			} else {
				// Recurse
				for _, child := range node.Children {
					categorizeNode(child)
				}
			}
		} else {
			ext := strings.ToLower(filepath.Ext(node.Name))
			baseName := strings.ToLower(node.Name)

			// Update extension stats
			if ext != "" {
				breakdown.ByExtension[ext] += node.Size
			}

			// Categorize by file type
			if ext == ".dex" {
				breakdown.DEX += node.Size
				breakdown.ByCategory["DEX Files"] += node.Size
			} else if ext == ".so" {
				breakdown.Libraries += node.Size
				breakdown.ByCategory["Native Libraries"] += node.Size
			} else if ext == ".png" || ext == ".jpg" || ext == ".jpeg" ||
				ext == ".gif" || ext == ".webp" || ext == ".xml" {
				breakdown.Resources += node.Size
				breakdown.ByCategory["Resources"] += node.Size
			} else if baseName == "androidmanifest.xml" || baseName == "resources.arsc" {
				breakdown.Resources += node.Size
				breakdown.ByCategory["Resources"] += node.Size
			} else {
				breakdown.Other += node.Size
				breakdown.ByCategory["Other"] += node.Size
			}
		}
	}

	for _, node := range nodes {
		categorizeNode(node)
	}

	return breakdown
}
