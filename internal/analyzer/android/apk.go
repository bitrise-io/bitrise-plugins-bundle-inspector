// Package android provides analyzers for Android artifacts.
package android

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
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
	return util.ValidateFileArtifact(path, ".apk")
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

	// Extract app icon
	iconData, err := util.ExtractIconFromZip(path, "apk")
	if err != nil {
		// Non-fatal, continue without icon
		iconData = ""
	}

	// Extract app metadata from manifest
	appName := ""
	packageName := ""
	version := ""
	if name, ok := manifest["app_name"].(string); ok {
		appName = name
	}
	if pkg, ok := manifest["package"].(string); ok {
		packageName = pkg
	}
	if ver, ok := manifest["version"].(string); ok {
		version = ver
	}

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:             path,
			Type:             types.ArtifactTypeAPK,
			Size:             info.Size(),
			UncompressedSize: uncompressedSize,
			AnalyzedAt:       time.Now(),
			IconData:         iconData,
			AppName:          appName,
			BundleID:         packageName,
			Version:          version,
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
		return manifest, fmt.Errorf("failed to open APK file: %w", err)
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

// categorizeAPKDirectory checks if a directory should be categorized as a whole.
// Returns true if categorized.
func categorizeAPKDirectory(node *types.FileNode, breakdown *types.SizeBreakdown) bool {
	dirName := strings.ToLower(node.Name)

	switch dirName {
	case "lib":
		breakdown.Libraries += node.Size
		breakdown.ByCategory["Native Libraries"] += node.Size
		return true
	case "res":
		breakdown.Resources += node.Size
		breakdown.ByCategory["Resources"] += node.Size
		return true
	case "assets":
		breakdown.Assets += node.Size
		breakdown.ByCategory["Assets"] += node.Size
		return true
	}

	return false
}

// categorizeAPKFile categorizes an individual file and updates the breakdown
func categorizeAPKFile(node *types.FileNode, breakdown *types.SizeBreakdown) {
	ext := util.GetLowerExtension(node.Name)
	baseName := strings.ToLower(node.Name)

	// Update extension stats
	if ext != "" {
		breakdown.ByExtension[ext] += node.Size
	}

	// Categorize by file type
	if ext == ".dex" {
		breakdown.DEX += node.Size
		breakdown.ByCategory["DEX Files"] += node.Size
		return
	}

	if ext == ".so" {
		breakdown.Libraries += node.Size
		breakdown.ByCategory["Native Libraries"] += node.Size
		return
	}

	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" ||
		ext == ".gif" || ext == ".webp" || ext == ".xml" {
		breakdown.Resources += node.Size
		breakdown.ByCategory["Resources"] += node.Size
		return
	}

	if baseName == "androidmanifest.xml" || baseName == "resources.arsc" {
		breakdown.Resources += node.Size
		breakdown.ByCategory["Resources"] += node.Size
		return
	}

	breakdown.Other += node.Size
	breakdown.ByCategory["Other"] += node.Size
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
			// Check if this directory should be categorized as a whole
			if categorizeAPKDirectory(node, &breakdown) {
				return
			}

			// Otherwise, recurse into children
			for _, child := range node.Children {
				categorizeNode(child)
			}
			return
		}

		// Categorize file
		categorizeAPKFile(node, &breakdown)
	}

	for _, node := range nodes {
		categorizeNode(node)
	}

	return breakdown
}
