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

// AABAnalyzer analyzes Android App Bundle files.
type AABAnalyzer struct{}

// NewAABAnalyzer creates a new AAB analyzer.
func NewAABAnalyzer() *AABAnalyzer {
	return &AABAnalyzer{}
}

// ValidateArtifact checks if the file is a valid AAB.
func (a *AABAnalyzer) ValidateArtifact(path string) error {
	return util.ValidateFileArtifact(path, ".aab")
}

// Analyze performs analysis on an AAB file.
func (a *AABAnalyzer) Analyze(ctx context.Context, path string) (*types.Report, error) {
	if err := a.ValidateArtifact(path); err != nil {
		return nil, err
	}

	// Get AAB file size
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat AAB: %w", err)
	}

	// Open AAB as ZIP
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open AAB: %w", err)
	}
	defer zipReader.Close()

	// Build file tree and calculate sizes
	fileTree, uncompressedSize := util.BuildZipFileTree(&zipReader.Reader)

	// Detect modules
	modules := detectModules(fileTree)

	// Create size breakdown
	sizeBreakdown := categorizeAABSizes(fileTree)

	// Find largest files
	largestFiles := util.FindLargestFiles(fileTree, 10)

	metadata := map[string]interface{}{
		"modules": modules,
	}

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:             path,
			Type:             types.ArtifactTypeAAB,
			Size:             info.Size(),
			UncompressedSize: uncompressedSize,
			AnalyzedAt:       time.Now(),
		},
		SizeBreakdown: sizeBreakdown,
		FileTree:      fileTree,
		LargestFiles:  largestFiles,
		Metadata:      metadata,
	}

	return report, nil
}

// detectModules identifies modules in the AAB
func detectModules(nodes []*types.FileNode) []string {
	modules := []string{}
	for _, node := range nodes {
		if node.IsDir {
			// AAB modules are typically at the root level
			modules = append(modules, node.Name)
		}
	}
	return modules
}

// categorizeAABDirectory checks if a directory should be categorized as a whole.
// Returns true if categorized.
func categorizeAABDirectory(node *types.FileNode, breakdown *types.SizeBreakdown) bool {
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
	case "dex":
		breakdown.DEX += node.Size
		breakdown.ByCategory["DEX Files"] += node.Size
		return true
	}

	return false
}

// categorizeAABFile categorizes an individual file and updates the breakdown
func categorizeAABFile(node *types.FileNode, breakdown *types.SizeBreakdown) {
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

	if baseName == "androidmanifest.xml" || baseName == "resources.pb" ||
		baseName == "bundleconfig.pb" {
		breakdown.Resources += node.Size
		breakdown.ByCategory["Resources"] += node.Size
		return
	}

	breakdown.Other += node.Size
	breakdown.ByCategory["Other"] += node.Size
}

// categorizeAABSizes creates a size breakdown for AAB
func categorizeAABSizes(nodes []*types.FileNode) types.SizeBreakdown {
	breakdown := types.SizeBreakdown{
		ByCategory:  make(map[string]int64),
		ByExtension: make(map[string]int64),
	}

	var categorizeNode func(node *types.FileNode, modulePrefix string)
	categorizeNode = func(node *types.FileNode, modulePrefix string) {
		if node.IsDir {
			// Track module sizes
			if modulePrefix == "" {
				modulePrefix = node.Name
				breakdown.ByCategory["Module: "+modulePrefix] = node.Size
			}

			// Check if this directory should be categorized as a whole
			if categorizeAABDirectory(node, &breakdown) {
				return
			}

			// Otherwise, recurse into children
			for _, child := range node.Children {
				categorizeNode(child, modulePrefix)
			}
			return
		}

		// Categorize file
		categorizeAABFile(node, &breakdown)
	}

	for _, node := range nodes {
		categorizeNode(node, "")
	}

	return breakdown
}
