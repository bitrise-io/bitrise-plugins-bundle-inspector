package ios

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// AppAnalyzer analyzes iOS .app bundles (uncompressed directories).
type AppAnalyzer struct{}

// NewAppAnalyzer creates a new .app analyzer.
func NewAppAnalyzer() *AppAnalyzer {
	return &AppAnalyzer{}
}

// ValidateArtifact checks if the path is a valid .app bundle.
func (a *AppAnalyzer) ValidateArtifact(path string) error {
	if !strings.HasSuffix(strings.ToLower(path), ".app") {
		return fmt.Errorf("path must have .app extension")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path must be a directory")
	}

	return nil
}

// Analyze performs analysis on a .app bundle directory.
func (a *AppAnalyzer) Analyze(ctx context.Context, path string) (*types.Report, error) {
	if err := a.ValidateArtifact(path); err != nil {
		return nil, err
	}

	// Analyze the .app bundle directory
	fileTree, totalSize, err := analyzeDirectory(path, "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze app bundle: %w", err)
	}

	// Analyze Mach-O binaries in file tree
	binaries := analyzeMachOBinaries(fileTree, path)

	// Discover frameworks
	frameworks, err := DiscoverFrameworks(path)
	if err != nil {
		// Not a critical error, just log it
		_ = err
	}

	// Convert binaries map to macho.BinaryInfo for dependency graph
	machoBinaries := make(map[string]*macho.BinaryInfo)
	for fwPath, binInfo := range binaries {
		machoBinaries[fwPath] = &macho.BinaryInfo{
			Architecture:    binInfo.Architecture,
			Architectures:   binInfo.Architectures,
			Type:            binInfo.Type,
			CodeSize:        binInfo.CodeSize,
			DataSize:        binInfo.DataSize,
			LinkedLibraries: binInfo.LinkedLibraries,
			RPaths:          binInfo.RPaths,
			HasDebugSymbols: binInfo.HasDebugSymbols,
		}
	}

	// Build dependency graph from binaries
	depGraph := macho.BuildDependencyGraph(machoBinaries)

	// Find main binary
	mainBinaryPath := findMainBinary(fileTree)

	// Detect unused frameworks
	var unusedFrameworks []string
	if mainBinaryPath != "" && len(depGraph) > 0 {
		unusedFrameworks = macho.DetectUnusedFrameworks(depGraph, mainBinaryPath)
	}

	// Parse asset catalogs
	assetCatalogs := parseAssetCatalogs(fileTree, path)

	// Create size breakdown
	sizeBreakdown := categorizeSizes(fileTree)

	// Find largest files
	largestFiles := findLargestFiles(fileTree, 10)

	// Prepare optimizations list
	var optimizations []types.Optimization

	// Add unused framework optimizations
	for _, fwPath := range unusedFrameworks {
		// Find framework info to get size
		var fwSize int64
		var fwName string
		for _, fw := range frameworks {
			if strings.Contains(fwPath, fw.Name) {
				fwSize = fw.Size
				fwName = fw.Name
				break
			}
		}

		if fwName != "" {
			optimizations = append(optimizations, types.Optimization{
				Category:    "frameworks",
				Severity:    "medium",
				Title:       fmt.Sprintf("Unused framework: %s", fwName),
				Description: "Framework is not linked by main binary or other frameworks",
				Action:      "Remove framework to reduce app size",
				Files:       []string{fwPath},
				Impact:      fwSize,
			})
		}
	}

	// Add optimization suggestions for oversized assets
	for _, catalog := range assetCatalogs {
		for _, asset := range catalog.LargestAssets {
			if asset.Size > 1*1024*1024 { // >1MB
				optimizations = append(optimizations, types.Optimization{
					Category:    "assets",
					Severity:    "low",
					Title:       fmt.Sprintf("Large asset: %s", asset.Name),
					Description: fmt.Sprintf("Asset is %s", util.FormatBytes(asset.Size)),
					Files:       []string{asset.Name},
					Impact:      asset.Size,
					Action:      "Consider compressing or resizing asset",
				})
			}
		}
	}

	// Convert frameworks to types.FrameworkInfo
	typedFrameworks := make([]*types.FrameworkInfo, len(frameworks))
	for i, fw := range frameworks {
		var binInfo *types.BinaryInfo
		if fw.BinaryInfo != nil {
			binInfo = &types.BinaryInfo{
				Architecture:    fw.BinaryInfo.Architecture,
				Architectures:   fw.BinaryInfo.Architectures,
				Type:            fw.BinaryInfo.Type,
				CodeSize:        fw.BinaryInfo.CodeSize,
				DataSize:        fw.BinaryInfo.DataSize,
				LinkedLibraries: fw.BinaryInfo.LinkedLibraries,
				RPaths:          fw.BinaryInfo.RPaths,
				HasDebugSymbols: fw.BinaryInfo.HasDebugSymbols,
			}
		}
		typedFrameworks[i] = &types.FrameworkInfo{
			Name:         fw.Name,
			Path:         fw.Path,
			Version:      fw.Version,
			Size:         fw.Size,
			BinaryInfo:   binInfo,
			Dependencies: fw.Dependencies,
		}
	}

	// Convert asset catalogs to types.AssetCatalogInfo
	typedAssetCatalogs := make([]*types.AssetCatalogInfo, len(assetCatalogs))
	for i, catalog := range assetCatalogs {
		largestAssets := make([]types.AssetInfo, len(catalog.LargestAssets))
		for j, asset := range catalog.LargestAssets {
			largestAssets[j] = types.AssetInfo{
				Name:  asset.Name,
				Type:  asset.Type,
				Scale: asset.Scale,
				Size:  asset.Size,
			}
		}
		typedAssetCatalogs[i] = &types.AssetCatalogInfo{
			Path:          catalog.Path,
			TotalSize:     catalog.TotalSize,
			AssetCount:    catalog.AssetCount,
			ByType:        catalog.ByType,
			ByScale:       catalog.ByScale,
			LargestAssets: largestAssets,
		}
	}

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:             path,
			Type:             types.ArtifactTypeApp,
			Size:             totalSize, // For .app, size is uncompressed
			UncompressedSize: totalSize,
			AnalyzedAt:       time.Now(),
		},
		SizeBreakdown:  sizeBreakdown,
		FileTree:       fileTree,
		LargestFiles:   largestFiles,
		Optimizations:  optimizations,
		Metadata: map[string]interface{}{
			"app_bundle":       filepath.Base(path),
			"is_directory":     true,
			"binaries":         binaries,
			"frameworks":       typedFrameworks,
			"dependency_graph": depGraph,
			"asset_catalogs":   typedAssetCatalogs,
		},
	}

	return report, nil
}
