// Package ios provides analyzers for iOS artifacts.
package ios

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
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

	// Analyze Mach-O binaries in file tree
	binaries := analyzeMachOBinaries(fileTree, appBundlePath)

	// Discover frameworks
	frameworks, err := DiscoverFrameworks(appBundlePath)
	if err != nil {
		log.Printf("Warning: Failed to discover frameworks: %v", err)
	}

	// Build dependency graph from binaries (no conversion needed - types are unified)
	depGraph := macho.BuildDependencyGraph(binaries)

	// Find main binary (typically the executable without extension at root)
	mainBinaryPath := findMainBinary(fileTree)

	// Detect unused frameworks
	var unusedFrameworks []string
	if mainBinaryPath != "" && len(depGraph) > 0 {
		unusedFrameworks = macho.DetectUnusedFrameworks(depGraph, mainBinaryPath)
	}

	// Parse asset catalogs
	assetCatalogs := parseAssetCatalogs(fileTree, appBundlePath)

	// Create size breakdown
	sizeBreakdown := categorizeSizes(fileTree)

	// Find largest files
	largestFiles := util.FindLargestFiles(fileTree, 10)

	// Prepare optimizations list
	var optimizations []types.Optimization

	// Generate symbol stripping optimizations
	symbolOpts := generateStripSymbolsOptimizations(binaries)
	for _, opt := range symbolOpts {
		optimizations = append(optimizations, *opt)
	}

	// Add unused framework optimizations
	frameworkOpts := GenerateUnusedFrameworkOptimizations(unusedFrameworks, frameworks)
	optimizations = append(optimizations, frameworkOpts...)

	// Add optimization suggestions for oversized assets
	assetOpts := GenerateLargeAssetOptimizations(assetCatalogs)
	optimizations = append(optimizations, assetOpts...)

	// Convert frameworks and asset catalogs to types
	typedFrameworks := ConvertFrameworksToTypes(frameworks)
	typedAssetCatalogs := ConvertAssetCatalogsToTypes(assetCatalogs)

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:             path,
			Type:             types.ArtifactTypeIPA,
			Size:             info.Size(),
			UncompressedSize: totalSize,
			AnalyzedAt:       time.Now(),
		},
		SizeBreakdown:  sizeBreakdown,
		FileTree:       fileTree,
		LargestFiles:   largestFiles,
		Optimizations:  optimizations,
		Metadata: map[string]interface{}{
			"app_bundle":       filepath.Base(appBundlePath),
			"binaries":         binaries,
			"frameworks":       typedFrameworks,
			"dependency_graph": depGraph,
			"asset_catalogs":   typedAssetCatalogs,
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

// isStickerExtensionBinary checks if a file path points to a stickers extension binary.
// Stickers extensions are primarily data containers (images/GIFs in .stickerpack)
// and don't have meaningful binary dependencies to analyze.
func isStickerExtensionBinary(filePath, rootPath string) bool {
	// Check if this is inside a .appex bundle
	if !strings.Contains(filePath, ".appex/") {
		return false
	}

	// Get the .appex directory path
	parts := strings.Split(filePath, ".appex/")
	if len(parts) < 2 {
		return false
	}
	appexDir := filepath.Join(rootPath, parts[0]+".appex")

	// Check if the .appex contains a .stickerpack subdirectory
	entries, err := os.ReadDir(appexDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), ".stickerpack") {
			return true
		}
	}

	return false
}

// analyzeMachOBinaries scans the file tree for Mach-O binaries and parses them.
func analyzeMachOBinaries(nodes []*types.FileNode, rootPath string) map[string]*types.BinaryInfo {
	binaries := make(map[string]*types.BinaryInfo)

	var walkNodes func(node *types.FileNode)
	walkNodes = func(node *types.FileNode) {
		if node.IsDir {
			for _, child := range node.Children {
				walkNodes(child)
			}
			return
		}

		fullPath := filepath.Join(rootPath, node.Path)

		// Detect Mach-O binaries by magic bytes
		if macho.IsMachO(fullPath) {
			// Skip stickers extension binaries - they're data containers, not functional binaries
			if isStickerExtensionBinary(node.Path, rootPath) {
				return
			}

			if info, err := macho.ParseMachO(fullPath); err == nil {
				// Convert internal BinaryInfo to types.BinaryInfo
				binaries[node.Path] = &types.BinaryInfo{
					Architecture:     info.Architecture,
					Architectures:    info.Architectures,
					Type:             info.Type,
					CodeSize:         info.CodeSize,
					DataSize:         info.DataSize,
					LinkedLibraries:  info.LinkedLibraries,
					RPaths:           info.RPaths,
					HasDebugSymbols:  info.HasDebugSymbols,
					DebugSymbolsSize: info.DebugSymbolsSize,
				}
			} else {
				// Graceful degradation: log warning, continue
				log.Printf("Warning: Failed to parse Mach-O: %s: %v", node.Path, err)
			}
		}
	}

	for _, node := range nodes {
		walkNodes(node)
	}

	return binaries
}

// findMainBinary identifies the main executable binary in the file tree.
// The main binary is typically at the root level and has no extension.
func findMainBinary(nodes []*types.FileNode) string {
	for _, node := range nodes {
		if !node.IsDir && filepath.Ext(node.Name) == "" {
			// Check if it's a Mach-O binary
			// Typically the main binary has the same name as the app bundle
			return node.Path
		}
	}
	return ""
}

// parseAssetCatalogs scans the file tree for .car files and parses them.
func parseAssetCatalogs(nodes []*types.FileNode, rootPath string) []*assets.AssetCatalogInfo {
	var catalogs []*assets.AssetCatalogInfo

	var walkNodes func(node *types.FileNode)
	walkNodes = func(node *types.FileNode) {
		if node.IsDir {
			for _, child := range node.Children {
				walkNodes(child)
			}
			return
		}

		if strings.HasSuffix(strings.ToLower(node.Name), ".car") {
			fullPath := filepath.Join(rootPath, node.Path)
			if catalog, err := assets.ParseAssetCatalog(fullPath); err == nil {
				catalogs = append(catalogs, catalog)
			} else {
				log.Printf("Warning: Failed to parse Assets.car: %s: %v", node.Path, err)
			}
		}
	}

	for _, node := range nodes {
		walkNodes(node)
	}

	return catalogs
}

// generateStripSymbolsOptimizations creates optimization recommendations for binaries with debug symbols.
func generateStripSymbolsOptimizations(binaries map[string]*types.BinaryInfo) []*types.Optimization {
	var optimizations []*types.Optimization

	for path, binary := range binaries {
		if binary.HasDebugSymbols && binary.DebugSymbolsSize > 0 {
			opt := &types.Optimization{
				Category:    "strip-symbols",
				Severity:    "high",
				Title:       fmt.Sprintf("Strip debug symbols from %s", filepath.Base(path)),
				Description: fmt.Sprintf("Binary contains debug symbols that can be removed. Symbol table size: %s", util.FormatBytes(binary.DebugSymbolsSize)),
				Impact:      binary.DebugSymbolsSize,
				Files:       []string{path},
				Action:      "Run 'strip -x' on this binary to remove debug symbols",
			}
			optimizations = append(optimizations, opt)
		}
	}

	// Sort by impact (largest first)
	sort.Slice(optimizations, func(i, j int) bool {
		return optimizations[i].Impact > optimizations[j].Impact
	})

	return optimizations
}
