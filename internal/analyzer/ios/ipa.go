// Package ios provides analyzers for iOS artifacts.
package ios

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/logger"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// IPAAnalyzer analyzes iOS IPA files.
type IPAAnalyzer struct {
	Logger logger.Logger
}

// NewIPAAnalyzer creates a new IPA analyzer.
func NewIPAAnalyzer(log logger.Logger) *IPAAnalyzer {
	if log == nil {
		log = logger.NewSilentLogger()
	}
	return &IPAAnalyzer{Logger: log}
}

// ValidateArtifact checks if the file is a valid IPA.
func (a *IPAAnalyzer) ValidateArtifact(path string) error {
	return util.ValidateFileArtifact(path, ".ipa")
}

// appBundleAnalysis holds the results of analyzing an app bundle
type appBundleAnalysis struct {
	appBundlePath    string
	fileTree         []*types.FileNode
	totalSize        int64
	binaries         map[string]*types.BinaryInfo
	frameworks       []*FrameworkInfo
	unusedFrameworks []string
	assetCatalogs    []*assets.AssetCatalogInfo
	appMetadata      *AppMetadata
	sizeBreakdown    types.SizeBreakdown
	largestFiles     []types.FileNode
}

// extractAndFindAppBundle extracts the IPA and finds the .app bundle
func (a *IPAAnalyzer) extractAndFindAppBundle(path string) (tempDir, appBundlePath string, err error) {
	tempDir, err = util.ExtractZip(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to extract IPA: %w", err)
	}

	appBundlePath, err = findAppBundle(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", "", err
	}

	return tempDir, appBundlePath, nil
}

// analyzeAppBundleContents performs comprehensive analysis of the app bundle
func (a *IPAAnalyzer) analyzeAppBundleContents(appBundlePath string) (*appBundleAnalysis, error) {
	// Analyze directory structure
	fileTree, totalSize, err := analyzeDirectory(appBundlePath, "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze app bundle: %w", err)
	}

	// Analyze binaries and frameworks
	binaries := analyzeMachOBinaries(fileTree, appBundlePath)
	frameworks, err := DiscoverFrameworks(appBundlePath)
	if err != nil {
		a.Logger.Warn("Failed to discover frameworks: %v", err)
	}

	// Analyze dependencies
	depGraph := macho.BuildDependencyGraph(binaries)
	mainBinaryPath := findMainBinary(fileTree)

	var unusedFrameworks []string
	if mainBinaryPath != "" && len(depGraph) > 0 {
		unusedFrameworks = macho.DetectUnusedFrameworks(depGraph, mainBinaryPath)
	}

	// Analyze assets
	assetCatalogs := parseAssetCatalogs(fileTree, appBundlePath, a.Logger)

	// Parse app metadata from Info.plist
	var appMetadata *AppMetadata
	infoPlistPath := filepath.Join(appBundlePath, "Info.plist")
	if parsedMetadata, err := ParseAppInfoPlist(infoPlistPath); err == nil {
		appMetadata = parsedMetadata
	}

	// Expand Mach-O binary segments as virtual children
	expandMachOSegments(fileTree, appBundlePath, a.Logger)

	return &appBundleAnalysis{
		appBundlePath:    appBundlePath,
		fileTree:         fileTree,
		totalSize:        totalSize,
		binaries:         binaries,
		frameworks:       frameworks,
		unusedFrameworks: unusedFrameworks,
		assetCatalogs:    assetCatalogs,
		appMetadata:      appMetadata,
		sizeBreakdown:    categorizeSizes(fileTree),
		largestFiles:     util.FindLargestFiles(fileTree, 10),
	}, nil
}

// generateAllOptimizations creates optimization suggestions from analysis results
func (a *IPAAnalyzer) generateAllOptimizations(analysis *appBundleAnalysis) []types.Optimization {
	var optimizations []types.Optimization

	// Symbol stripping optimizations
	symbolOpts := generateStripSymbolsOptimizations(analysis.binaries)
	for _, opt := range symbolOpts {
		optimizations = append(optimizations, *opt)
	}

	// Unused framework optimizations
	frameworkOpts := GenerateUnusedFrameworkOptimizations(
		analysis.unusedFrameworks,
		analysis.frameworks,
	)
	optimizations = append(optimizations, frameworkOpts...)

	// Oversized asset optimizations
	assetOpts := GenerateLargeAssetOptimizations(analysis.assetCatalogs)
	optimizations = append(optimizations, assetOpts...)

	return optimizations
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

	// Extract and locate app bundle
	tempDir, appBundlePath, err := a.extractAndFindAppBundle(path)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// Analyze app bundle contents
	analysis, err := a.analyzeAppBundleContents(appBundlePath)
	if err != nil {
		return nil, err
	}

	// Generate optimizations
	optimizations := a.generateAllOptimizations(analysis)

	// Build metadata map
	metadata := map[string]interface{}{
		"app_bundle":       filepath.Base(analysis.appBundlePath),
		"binaries":         analysis.binaries,
		"frameworks":       ConvertFrameworksToTypes(analysis.frameworks),
		"dependency_graph": macho.BuildDependencyGraph(analysis.binaries),
		"asset_catalogs":   ConvertAssetCatalogsToTypes(analysis.assetCatalogs),
		"platform":         "iOS",
	}

	// Add app metadata if available
	if analysis.appMetadata != nil {
		if analysis.appMetadata.AppName != "" {
			metadata["app_name"] = analysis.appMetadata.AppName
		}
		if analysis.appMetadata.BundleID != "" {
			metadata["bundle_id"] = analysis.appMetadata.BundleID
		}
		if analysis.appMetadata.Version != "" {
			metadata["version"] = analysis.appMetadata.Version
		}
		if analysis.appMetadata.BuildVersion != "" {
			metadata["build_version"] = analysis.appMetadata.BuildVersion
		}
		if analysis.appMetadata.MinOSVersion != "" {
			metadata["min_os_version"] = analysis.appMetadata.MinOSVersion
		}
	}

	// Extract app icon with Info.plist-guided search
	var iconHints *util.IconSearchHints
	if analysis.appMetadata != nil && len(analysis.appMetadata.IconNames) > 0 {
		iconHints = &util.IconSearchHints{PlistIconNames: analysis.appMetadata.IconNames}
	}
	iconData, err := util.ExtractIconFromZipWithHints(path, "ipa", iconHints)
	if err == nil && iconData != "" {
		a.Logger.Info("Icon extracted from loose file in archive")
	} else {
		if err != nil {
			a.Logger.Debug("Loose icon extraction failed: %v, trying Assets.car fallback", err)
		}
		// Fallback: try extracting icon from Assets.car
		if carIcon := tryExtractIconFromAssetsCar(ctx, appBundlePath, analysis.appMetadata, analysis.assetCatalogs); carIcon != "" {
			iconData = carIcon
			a.Logger.Info("Icon extracted from Assets.car")
		}
		// Continue without icon - it's not critical
	}

	// Build artifact info
	artifactInfo := types.ArtifactInfo{
		Path:             path,
		Type:             types.ArtifactTypeIPA,
		Size:             info.Size(),
		UncompressedSize: analysis.totalSize,
		AnalyzedAt:       time.Now(),
		IconData:         iconData,
	}

	// Add app metadata to artifact info if available
	if analysis.appMetadata != nil {
		artifactInfo.AppName = analysis.appMetadata.AppName
		artifactInfo.BundleID = analysis.appMetadata.BundleID
		artifactInfo.Version = analysis.appMetadata.Version
	}

	// Build final report
	report := &types.Report{
		ArtifactInfo:  artifactInfo,
		SizeBreakdown: analysis.sizeBreakdown,
		FileTree:      analysis.fileTree,
		LargestFiles:  analysis.largestFiles,
		Optimizations: optimizations,
		Metadata:      metadata,
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

// categorizeIOSDirectory checks if a directory should be categorized as a whole
// and updates the breakdown accordingly. Returns true if categorized.
func categorizeIOSDirectory(node *types.FileNode, breakdown *types.SizeBreakdown) bool {
	dirName := strings.ToLower(node.Name)

	if strings.HasSuffix(dirName, ".framework") || dirName == "frameworks" {
		breakdown.Frameworks += node.Size
		breakdown.ByCategory["Frameworks"] += node.Size
		return true
	}

	return false
}

// categorizeIOSFile categorizes an individual file and updates the breakdown
func categorizeIOSFile(node *types.FileNode, breakdown *types.SizeBreakdown) {
	ext := util.GetLowerExtension(node.Name)
	baseName := strings.ToLower(node.Name)

	// Update extension stats
	if ext != "" {
		breakdown.ByExtension[ext] += node.Size
	}

	// Categorize by type
	if baseName == filepath.Base(node.Path) && ext == "" {
		// Likely the main executable
		breakdown.Executable += node.Size
		breakdown.ByCategory["Executable"] += node.Size
		return
	}

	if ext == ".dylib" || ext == ".a" || ext == ".so" {
		breakdown.Libraries += node.Size
		breakdown.ByCategory["Libraries"] += node.Size
		return
	}

	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" ||
		ext == ".car" || ext == ".pdf" || ext == ".svg" {
		breakdown.Assets += node.Size
		breakdown.ByCategory["Assets"] += node.Size
		return
	}

	if ext == ".nib" || ext == ".storyboard" || ext == ".storyboardc" ||
		ext == ".strings" || ext == ".plist" || ext == ".json" {
		breakdown.Resources += node.Size
		breakdown.ByCategory["Resources"] += node.Size
		return
	}

	breakdown.Other += node.Size
	breakdown.ByCategory["Other"] += node.Size
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
			// Check if this directory should be categorized as a whole
			if categorizeIOSDirectory(node, &breakdown) {
				return
			}

			// Otherwise, recurse into children
			for _, child := range node.Children {
				categorizeNode(child)
			}
			return
		}

		// Categorize file
		categorizeIOSFile(node, &breakdown)
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

			info, err := macho.ParseMachO(fullPath)
			if err != nil {
				// Graceful degradation: log warning, continue
				return
			}

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
		}
	}

	for _, node := range nodes {
		walkNodes(node)
	}

	return binaries
}

// findMainBinary identifies the main executable binary in the file tree.
// The main binary is typically at the root level and has no extension.
// It filters out metadata files like PkgInfo and selects the largest candidate.
func findMainBinary(nodes []*types.FileNode) string {
	var candidates []struct {
		path string
		size int64
	}

	// Find all files without extension at root level
	for _, node := range nodes {
		if !node.IsDir && filepath.Ext(node.Name) == "" {
			// Skip known metadata files
			if isMetadataFile(node.Name) {
				continue
			}
			candidates = append(candidates, struct {
				path string
				size int64
			}{node.Path, node.Size})
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Return the largest file (main executable is typically largest)
	// This handles cases where multiple extensionless files exist
	maxIdx := 0
	for i := 1; i < len(candidates); i++ {
		if candidates[i].size > candidates[maxIdx].size {
			maxIdx = i
		}
	}

	return candidates[maxIdx].path
}

// isMetadataFile checks if a filename is a known iOS metadata file that should
// not be treated as the main executable.
func isMetadataFile(name string) bool {
	metadataFiles := []string{
		"PkgInfo",
		"CodeResources",
		"_CodeSignature",
		"embedded.mobileprovision",
	}
	for _, metadata := range metadataFiles {
		if name == metadata {
			return true
		}
	}
	return false
}

// parseAssetCatalogs scans the file tree for .car files and parses them.
// It also expands assets as virtual children in the file tree.
func parseAssetCatalogs(nodes []*types.FileNode, rootPath string, log logger.Logger) []*assets.AssetCatalogInfo {
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
			catalog, err := assets.ParseAssetCatalog(fullPath)
			if err != nil {
				log.Warn("Failed to parse Assets.car %s: %v", node.Path, err)
				return
			}

			// Store the full relative path for proper virtual path generation
			catalog.Path = node.Path

			// Expand assets as virtual children of the .car file node
			virtualChildren := assets.ExpandAssetsAsChildren(catalog, node.Path)
			if len(virtualChildren) > 0 {
				node.Children = virtualChildren
			}

			catalogs = append(catalogs, catalog)
		}
	}

	for _, node := range nodes {
		walkNodes(node)
	}

	return catalogs
}

// expandMachOSegments walks the file tree and expands Mach-O binaries
// to show their segments and sections as virtual children.
func expandMachOSegments(nodes []*types.FileNode, rootPath string, log logger.Logger) {
	var walkNodes func(node *types.FileNode)
	walkNodes = func(node *types.FileNode) {
		if node.IsDir {
			for _, child := range node.Children {
				walkNodes(child)
			}
			return
		}

		// Skip already virtual nodes (e.g., assets from .car files)
		if node.IsVirtual {
			return
		}

		fullPath := filepath.Join(rootPath, node.Path)

		// Check if this is a Mach-O binary
		if !macho.IsMachO(fullPath) {
			return
		}

		// Parse segments from the binary
		segments, err := macho.ParseSegments(fullPath)
		if err != nil {
			log.Warn("Failed to parse Mach-O segments for %s: %v", node.Path, err)
			return
		}

		// Expand segments as virtual children
		virtualChildren := macho.ExpandSegmentsAsChildren(segments, node.Path)
		if len(virtualChildren) > 0 {
			node.Children = virtualChildren
		}
	}

	for _, node := range nodes {
		walkNodes(node)
	}
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

// tryExtractIconFromAssetsCar attempts to extract an app icon from an Assets.car file.
// Returns a base64 data URI string, or empty string if extraction fails.
func tryExtractIconFromAssetsCar(ctx context.Context, appBundlePath string, appMetadata *AppMetadata, assetCatalogs []*assets.AssetCatalogInfo) string {
	carPath := filepath.Join(appBundlePath, "Assets.car")
	if _, err := os.Stat(carPath); err != nil {
		return ""
	}

	var iconNames []string
	if appMetadata != nil {
		iconNames = appMetadata.IconNames
	}

	var catalogAssets []assets.AssetInfo
	for _, cat := range assetCatalogs {
		catalogAssets = append(catalogAssets, cat.Assets...)
	}

	carIcon, err := assets.ExtractIconFromCar(ctx, carPath, iconNames, catalogAssets)
	if err != nil || len(carIcon) == 0 {
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString(carIcon)
	return fmt.Sprintf("data:image/png;base64,%s", encoded)
}
