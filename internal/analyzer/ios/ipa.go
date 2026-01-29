// Package ios provides analyzers for iOS artifacts.
package ios

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// IPAAnalyzer analyzes iOS IPA files.
type IPAAnalyzer struct{}

// NewIPAAnalyzer creates a new IPA analyzer.
func NewIPAAnalyzer() *IPAAnalyzer {
	return &IPAAnalyzer{}
}

// ValidateArtifact checks if the file is a valid IPA.
func (a *IPAAnalyzer) ValidateArtifact(path string) error {
	if !strings.HasSuffix(strings.ToLower(path), ".ipa") {
		return fmt.Errorf("file must have .ipa extension")
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

// Analyze performs analysis on an IPA file.
func (a *IPAAnalyzer) Analyze(ctx context.Context, path string) (*types.Report, error) {
	if err := a.ValidateArtifact(path); err != nil {
		return nil, err
	}

	// Get IPA file size
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat IPA: %w", err)
	}

	// Extract IPA
	tempDir, err := util.ExtractZip(path)
	if err != nil {
		return nil, fmt.Errorf("failed to extract IPA: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Find .app bundle
	appBundlePath, err := findAppBundle(tempDir)
	if err != nil {
		return nil, err
	}

	// Analyze the .app bundle
	fileTree, totalSize, err := analyzeDirectory(appBundlePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze app bundle: %w", err)
	}

	// Create size breakdown
	sizeBreakdown := categorizeSizes(fileTree)

	// Find largest files
	largestFiles := findLargestFiles(fileTree, 10)

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:             path,
			Type:             types.ArtifactTypeIPA,
			Size:             info.Size(),
			UncompressedSize: totalSize,
			AnalyzedAt:       time.Now(),
		},
		SizeBreakdown: sizeBreakdown,
		FileTree:      fileTree,
		LargestFiles:  largestFiles,
		Metadata: map[string]interface{}{
			"app_bundle": filepath.Base(appBundlePath),
		},
	}

	return report, nil
}

// findAppBundle locates the .app bundle within the extracted IPA.
func findAppBundle(root string) (string, error) {
	var appPath string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasSuffix(path, ".app") {
			appPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if appPath == "" {
		return "", fmt.Errorf("no .app bundle found in IPA")
	}

	return appPath, nil
}

// analyzeDirectory recursively analyzes a directory and builds a file tree.
func analyzeDirectory(root, basePath string) ([]*types.FileNode, int64, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, 0, err
	}

	var nodes []*types.FileNode
	var totalSize int64

	for _, entry := range entries {
		fullPath := filepath.Join(root, entry.Name())
		relativePath := filepath.Join(basePath, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			// Recursively analyze subdirectory
			children, dirSize, err := analyzeDirectory(fullPath, relativePath)
			if err != nil {
				continue
			}

			node := &types.FileNode{
				Path:     relativePath,
				Name:     entry.Name(),
				Size:     dirSize,
				IsDir:    true,
				Children: children,
			}
			nodes = append(nodes, node)
			totalSize += dirSize
		} else {
			// Regular file
			node := &types.FileNode{
				Path:  relativePath,
				Name:  entry.Name(),
				Size:  info.Size(),
				IsDir: false,
			}
			nodes = append(nodes, node)
			totalSize += info.Size()
		}
	}

	return nodes, totalSize, nil
}

// categorizeSizes creates a size breakdown by category.
func categorizeSizes(nodes []*types.FileNode) types.SizeBreakdown {
	breakdown := types.SizeBreakdown{
		ByCategory:  make(map[string]int64),
		ByExtension: make(map[string]int64),
	}

	var categorizeNode func(node *types.FileNode)
	categorizeNode = func(node *types.FileNode) {
		if node.IsDir {
			// Categorize by directory name
			dirName := strings.ToLower(node.Name)

			if strings.HasSuffix(dirName, ".framework") {
				breakdown.Frameworks += node.Size
				breakdown.ByCategory["Frameworks"] += node.Size
			} else if dirName == "frameworks" {
				breakdown.Frameworks += node.Size
				breakdown.ByCategory["Frameworks"] += node.Size
			} else {
				// Recurse into children
				for _, child := range node.Children {
					categorizeNode(child)
				}
			}
		} else {
			// Categorize by file
			ext := strings.ToLower(filepath.Ext(node.Name))
			baseName := strings.ToLower(node.Name)

			// Update extension stats
			if ext != "" {
				breakdown.ByExtension[ext] += node.Size
			}

			// Categorize
			if baseName == filepath.Base(node.Path) && ext == "" {
				// Likely the main executable
				breakdown.Executable += node.Size
				breakdown.ByCategory["Executable"] += node.Size
			} else if ext == ".dylib" || ext == ".a" || ext == ".so" {
				breakdown.Libraries += node.Size
				breakdown.ByCategory["Libraries"] += node.Size
			} else if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" ||
					  ext == ".car" || ext == ".pdf" || ext == ".svg" {
				breakdown.Assets += node.Size
				breakdown.ByCategory["Assets"] += node.Size
			} else if ext == ".nib" || ext == ".storyboard" || ext == ".storyboardc" ||
					  ext == ".strings" || ext == ".plist" || ext == ".json" {
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

// findLargestFiles returns the N largest files from the tree.
func findLargestFiles(nodes []*types.FileNode, n int) []types.FileNode {
	var files []types.FileNode

	var collectFiles func(node *types.FileNode)
	collectFiles = func(node *types.FileNode) {
		if node.IsDir {
			for _, child := range node.Children {
				collectFiles(child)
			}
		} else {
			files = append(files, *node)
		}
	}

	for _, node := range nodes {
		collectFiles(node)
	}

	// Sort by size descending
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	// Return top N
	if len(files) > n {
		files = files[:n]
	}

	return files
}
