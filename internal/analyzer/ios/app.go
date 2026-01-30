package ios

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/logger"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// AppAnalyzer analyzes iOS .app bundles (uncompressed directories).
type AppAnalyzer struct {
	Logger logger.Logger
}

// NewAppAnalyzer creates a new .app analyzer.
func NewAppAnalyzer(log logger.Logger) *AppAnalyzer {
	if log == nil {
		log = logger.NewSilentLogger()
	}
	return &AppAnalyzer{Logger: log}
}

// ValidateArtifact checks if the path is a valid .app bundle.
func (a *AppAnalyzer) ValidateArtifact(path string) error {
	return util.ValidateDirectoryArtifact(path, ".app")
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

	// Build dependency graph from binaries (no conversion needed - types are unified)
	depGraph := macho.BuildDependencyGraph(binaries)

	// Find main binary
	mainBinaryPath := findMainBinary(fileTree)

	// Detect unused frameworks
	var unusedFrameworks []string
	if mainBinaryPath != "" && len(depGraph) > 0 {
		unusedFrameworks = macho.DetectUnusedFrameworks(depGraph, mainBinaryPath)
	}

	// Parse asset catalogs
	assetCatalogs := parseAssetCatalogs(fileTree, path, a.Logger)

	// Create size breakdown
	sizeBreakdown := categorizeSizes(fileTree)

	// Find largest files
	largestFiles := util.FindLargestFiles(fileTree, 10)

	// Prepare optimizations list
	var optimizations []types.Optimization

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
