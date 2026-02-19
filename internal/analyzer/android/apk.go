// Package android provides analyzers for Android artifacts.
package android

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/shogo82148/androidbinary/apk"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/android/dex"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/jsbundle"
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

	// Parse DEX files and create virtual tree
	dexTree, totalDEXSize, err := dex.ParseAndMerge(path, fileTree)
	if err != nil {
		// Non-fatal: keep original .dex files if parsing fails
		fmt.Fprintf(os.Stderr, "DEX parsing failed: %v\n", err)
	} else {
		// Replace individual .dex files with virtual Dex/ directory
		fileTree = dex.ReplaceDEXFilesWithVirtual(fileTree, dexTree)
	}

	// Detect JS bundle format for React Native apps
	jsBundleInfo := detectJSBundleInZip(path, fileTree)

	// Create size breakdown (after DEX replacement)
	sizeBreakdown := categorizeAPKSizes(fileTree)

	// Update DEX size if we parsed it
	if totalDEXSize > 0 {
		sizeBreakdown.DEX = totalDEXSize
	}

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

	// Add JS bundle info to metadata if detected (React Native)
	for k, v := range jsBundleInfo {
		manifest[k] = v
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

	// Try to parse APK using androidbinary library
	pkg, err := apk.OpenFile(apkPath)
	if err != nil {
		// If full parsing fails (e.g., incomplete APK, missing resources),
		// fall back to basic manifest detection
		zipFile, zipErr := zip.OpenReader(apkPath)
		if zipErr != nil {
			return manifest, fmt.Errorf("failed to open APK: %w", zipErr)
		}
		defer zipFile.Close()

		// Just detect manifest presence
		for _, f := range zipFile.File {
			if f.Name == "AndroidManifest.xml" {
				manifest["has_manifest"] = true
				break
			}
		}
		return manifest, nil
	}
	defer pkg.Close()

	// Extract package information
	manifest["has_manifest"] = true
	manifest["package"] = pkg.PackageName()

	// Extract version information
	m := pkg.Manifest()
	if versionName, err := m.VersionName.String(); err == nil && versionName != "" {
		manifest["version"] = versionName
	}
	if versionCode, err := m.VersionCode.Int32(); err == nil && versionCode > 0 {
		manifest["version_code"] = fmt.Sprintf("%d", versionCode)
	}

	// Extract app name (label)
	label, err := pkg.Label(nil) // nil for default locale
	if err == nil && label != "" {
		manifest["app_name"] = label
	}

	return manifest, nil
}

// categorizeAPKDirectory checks if a directory should be categorized as a whole.
// Returns true if categorized.
func categorizeAPKDirectory(node *types.FileNode, breakdown *types.SizeBreakdown) bool {
	dirName := strings.ToLower(node.Name)

	switch dirName {
	case "dex":
		// Virtual DEX directory with parsed classes
		breakdown.DEX += node.Size
		breakdown.ByCategory["DEX Files"] += node.Size
		return true
	case "lib":
		breakdown.Libraries += node.Size
		breakdown.ByCategory["Native Libraries"] += node.Size
		return true
	case "res":
		breakdown.Resources += node.Size
		breakdown.ByCategory["Resources"] += node.Size
		return true
	case "assets":
		// Check for JS bundles inside assets (React Native)
		jsSize := collectJSBundleSize(node.Children)
		if jsSize > 0 {
			breakdown.JavaScript += jsSize
			breakdown.ByCategory["JavaScript"] += jsSize
			nonJS := node.Size - jsSize
			if nonJS > 0 {
				breakdown.Assets += nonJS
				breakdown.ByCategory["Assets"] += nonJS
			}
		} else {
			breakdown.Assets += node.Size
			breakdown.ByCategory["Assets"] += node.Size
		}
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

// collectJSBundleSize recursively finds JS bundle files and returns their total size.
func collectJSBundleSize(nodes []*types.FileNode) int64 {
	var size int64
	for _, node := range nodes {
		if !node.IsDir && jsbundle.IsJSBundleFilename(node.Name) {
			size += node.Size
		} else if node.IsDir {
			size += collectJSBundleSize(node.Children)
		}
	}
	return size
}

// findJSBundleNode walks a file tree to find the first JS bundle file node.
func findJSBundleNode(nodes []*types.FileNode) *types.FileNode {
	for _, node := range nodes {
		if !node.IsDir && jsbundle.IsJSBundleFilename(node.Name) {
			return node
		}
		if node.IsDir {
			if found := findJSBundleNode(node.Children); found != nil {
				return found
			}
		}
	}
	return nil
}

// detectJSBundleInZip detects a JS bundle inside a ZIP archive and identifies its format.
// Returns metadata about the bundle, or nil if no JS bundle is found.
func detectJSBundleInZip(archivePath string, fileTree []*types.FileNode) map[string]interface{} {
	bundleNode := findJSBundleNode(fileTree)
	if bundleNode == nil {
		return nil
	}

	result := map[string]interface{}{
		"is_react_native": true,
		"js_bundle_path":  bundleNode.Path,
		"js_bundle_size":  bundleNode.Size,
	}

	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return result
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		if f.Name == bundleNode.Path {
			rc, err := f.Open()
			if err != nil {
				break
			}
			buf := make([]byte, 512)
			n, _ := io.ReadAtLeast(rc, buf, 4)
			rc.Close()
			if n >= 4 {
				format, err := jsbundle.DetectFormat(bytes.NewReader(buf[:n]))
				if err == nil {
					result["js_bundle_format"] = string(format)
				}
			}
			break
		}
	}

	return result
}
