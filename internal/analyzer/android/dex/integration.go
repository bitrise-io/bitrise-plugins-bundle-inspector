package dex

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// DetectDEXFiles finds all DEX files in the file tree.
func DetectDEXFiles(fileTree []*types.FileNode) []string {
	dexFiles := make([]string, 0)

	var walk func([]*types.FileNode)
	walk = func(nodes []*types.FileNode) {
		for _, node := range nodes {
			if !node.IsDir && strings.HasSuffix(node.Name, ".dex") {
				dexFiles = append(dexFiles, node.Path)
			}
			if node.IsDir && node.Children != nil {
				walk(node.Children)
			}
		}
	}

	walk(fileTree)

	// Sort to ensure consistent processing order
	sort.Strings(dexFiles)
	return dexFiles
}

// ParseAndMerge parses all DEX files from an APK/AAB and merges them into a virtual tree.
func ParseAndMerge(archivePath string, fileTree []*types.FileNode) (*types.FileNode, int64, error) {
	// 1. Detect all DEX files
	dexFiles := DetectDEXFiles(fileTree)
	if len(dexFiles) == 0 {
		return nil, 0, fmt.Errorf("no DEX files found")
	}

	// 2. Calculate total DEX file size from file tree
	totalDEXSize := int64(0)
	for _, dexPath := range dexFiles {
		node := findNodeByPath(fileTree, dexPath)
		if node != nil {
			totalDEXSize += node.Size
		}
	}

	// 3. Extract and parse DEX files
	tempDir, err := os.MkdirTemp("", "dex-extract-*")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	mergedInfo, err := extractAndMergeDEXFiles(archivePath, dexFiles, tempDir)
	if err != nil {
		return nil, 0, err
	}

	// 4. Build virtual tree
	dexTree := BuildVirtualDEXTree(mergedInfo, totalDEXSize)

	return dexTree, totalDEXSize, nil
}

// extractAndMergeDEXFiles extracts DEX files from archive and parses them.
func extractAndMergeDEXFiles(archivePath string, dexPaths []string, tempDir string) (*types.MergedDEXInfo, error) {
	// Open archive
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer reader.Close()

	allClasses := make([]types.DexClass, 0)
	totalPrivateSize := int64(0)
	totalFileSize := int64(0)

	for _, dexPath := range dexPaths {
		// Find DEX file in archive
		var dexFile *zip.File
		for _, f := range reader.File {
			if f.Name == dexPath {
				dexFile = f
				break
			}
		}

		if dexFile == nil {
			continue
		}

		// Extract to temp file
		tempPath := filepath.Join(tempDir, filepath.Base(dexPath))
		if err := extractFile(dexFile, tempPath); err != nil {
			continue
		}

		// Parse DEX file
		dexInfo, err := ParseDEXFile(tempPath)
		if err != nil {
			// Skip DEX files that fail to parse
			continue
		}

		// Merge classes
		for _, class := range dexInfo.Classes {
			// Update source DEX to use original path
			class.SourceDEX = dexPath
			allClasses = append(allClasses, class)
		}

		totalPrivateSize += dexInfo.TotalPrivateSize
		totalFileSize += dexInfo.TotalFileSize
	}

	return &types.MergedDEXInfo{
		Classes:          allClasses,
		TotalPrivateSize: totalPrivateSize,
		TotalFileSize:    totalFileSize,
		DEXFileCount:     len(dexPaths),
	}, nil
}

// extractFile extracts a file from a zip archive.
func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// findNodeByPath finds a node in the file tree by its path.
func findNodeByPath(nodes []*types.FileNode, path string) *types.FileNode {
	for _, node := range nodes {
		if node.Path == path {
			return node
		}
		if node.IsDir && node.Children != nil {
			if found := findNodeByPath(node.Children, path); found != nil {
				return found
			}
		}
	}
	return nil
}

// ReplaceDEXFilesWithVirtual replaces individual .dex files with a virtual Dex/ directory.
func ReplaceDEXFilesWithVirtual(fileTree []*types.FileNode, dexTree *types.FileNode) []*types.FileNode {
	result := make([]*types.FileNode, 0, len(fileTree))
	dexAdded := false

	for _, node := range fileTree {
		// Skip individual .dex files
		if !node.IsDir && strings.HasSuffix(node.Name, ".dex") {
			if !dexAdded {
				// Add virtual Dex/ directory once
				result = append(result, dexTree)
				dexAdded = true
			}
			continue
		}
		result = append(result, node)
	}

	return result
}
