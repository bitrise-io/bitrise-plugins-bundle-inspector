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

// AABAnalyzer analyzes Android App Bundle files.
type AABAnalyzer struct{}

// NewAABAnalyzer creates a new AAB analyzer.
func NewAABAnalyzer() *AABAnalyzer {
	return &AABAnalyzer{}
}

// ValidateArtifact checks if the file is a valid AAB.
func (a *AABAnalyzer) ValidateArtifact(path string) error {
	if !strings.HasSuffix(strings.ToLower(path), ".aab") {
		return fmt.Errorf("file must have .aab extension")
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

// categorizeAABSizes creates a size breakdown for AAB
func categorizeAABSizes(nodes []*types.FileNode) types.SizeBreakdown {
	breakdown := types.SizeBreakdown{
		ByCategory:  make(map[string]int64),
		ByExtension: make(map[string]int64),
	}

	var categorizeNode func(node *types.FileNode, modulePrefix string)
	categorizeNode = func(node *types.FileNode, modulePrefix string) {
		if node.IsDir {
			dirName := strings.ToLower(node.Name)

			// Track module sizes
			if modulePrefix == "" {
				modulePrefix = node.Name
				breakdown.ByCategory["Module: "+modulePrefix] = node.Size
			}

			// Categorize by directory within module
			if dirName == "lib" {
				breakdown.Libraries += node.Size
				breakdown.ByCategory["Native Libraries"] += node.Size
			} else if dirName == "res" {
				breakdown.Resources += node.Size
				breakdown.ByCategory["Resources"] += node.Size
			} else if dirName == "assets" {
				breakdown.Assets += node.Size
				breakdown.ByCategory["Assets"] += node.Size
			} else if dirName == "dex" {
				breakdown.DEX += node.Size
				breakdown.ByCategory["DEX Files"] += node.Size
			} else {
				// Recurse
				for _, child := range node.Children {
					categorizeNode(child, modulePrefix)
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
			} else if baseName == "androidmanifest.xml" || baseName == "resources.pb" ||
				baseName == "bundleconfig.pb" {
				breakdown.Resources += node.Size
				breakdown.ByCategory["Resources"] += node.Size
			} else {
				breakdown.Other += node.Size
				breakdown.ByCategory["Other"] += node.Size
			}
		}
	}

	for _, node := range nodes {
		categorizeNode(node, "")
	}

	return breakdown
}
