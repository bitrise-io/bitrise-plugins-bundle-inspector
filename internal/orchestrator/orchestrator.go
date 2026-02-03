// Package orchestrator coordinates the analysis workflow
package orchestrator

import (
	"context"
	"fmt"
	"os"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/detector"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/logger"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// Orchestrator coordinates the analysis workflow
type Orchestrator struct {
	IncludeDuplicates bool
	Logger            logger.Logger
}

// New creates a new orchestrator with default settings
func New() *Orchestrator {
	return &Orchestrator{
		IncludeDuplicates: true,
		Logger:            logger.NewDefaultLogger(os.Stderr, logger.LevelInfo),
	}
}

// RunAnalysis performs a complete analysis of an artifact
func (o *Orchestrator) RunAnalysis(ctx context.Context, artifactPath string) (*types.Report, error) {
	// Create analyzer
	a, err := analyzer.NewAnalyzer(artifactPath, o.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}

	// Perform initial analysis
	report, err := a.Analyze(ctx, artifactPath)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Run duplicate detection and additional optimizations if enabled
	if o.IncludeDuplicates {
		if err := o.runDetectors(report, artifactPath); err != nil {
			// Log warning but don't fail
			o.Logger.Warn("detector execution had issues: %v", err)
		}
	}

	// Generate optimization recommendations
	report.Optimizations = o.generateOptimizations(report)
	report.TotalSavings = calculateTotalSavings(report)

	// Add Git/CI metadata if available
	o.enrichWithCIMetadata(report)

	return report, nil
}

// runDetectors executes duplicate detection and additional optimization detectors
func (o *Orchestrator) runDetectors(report *types.Report, artifactPath string) error {
	// Extract artifact for duplicate detection
	extractPath, shouldCleanup, err := o.extractArtifact(report.ArtifactInfo.Type, artifactPath)
	if err != nil {
		return fmt.Errorf("failed to extract artifact: %w", err)
	}

	if shouldCleanup {
		defer os.RemoveAll(extractPath)
	}

	if extractPath == "" {
		return nil // Nothing to analyze
	}

	// Run duplicate detection for files
	dupDetector := detector.NewDuplicateDetector()
	duplicates, err := dupDetector.DetectDuplicates(extractPath)
	if err != nil {
		o.Logger.Warn("duplicate detection failed: %v", err)
	} else {
		report.Duplicates = duplicates
	}

	// Run asset duplicate detection for .car files
	assetDuplicates := o.detectAssetDuplicates(report)
	if len(assetDuplicates) > 0 {
		report.Duplicates = append(report.Duplicates, assetDuplicates...)
	}

	// Run additional detectors
	o.runAdditionalDetectors(report, extractPath)

	return nil
}

// extractArtifact extracts the artifact if needed and returns the path and cleanup flag
func (o *Orchestrator) extractArtifact(artifactType types.ArtifactType, artifactPath string) (extractPath string, shouldCleanup bool, err error) {
	switch artifactType {
	case types.ArtifactTypeIPA, types.ArtifactTypeAPK, types.ArtifactTypeAAB:
		extractPath, err = util.ExtractZip(artifactPath)
		if err != nil {
			return "", false, fmt.Errorf("failed to extract zip: %w", err)
		}
		return extractPath, true, nil

	case types.ArtifactTypeApp:
		// .app bundles are already directories, use them directly
		return artifactPath, false, nil

	default:
		return "", false, nil
	}
}

// runAdditionalDetectors runs all additional optimization detectors
func (o *Orchestrator) runAdditionalDetectors(report *types.Report, extractPath string) {
	detectors := []detector.Detector{
		detector.NewImageOptimizationDetector(),
		detector.NewLooseImagesDetector(),
		detector.NewUnnecessaryFilesDetector(),
	}

	// Add iOS-specific detectors
	if o.isIOSArtifact(report.ArtifactInfo.Type) {
		detectors = append(detectors, detector.NewSmallFilesDetector())
	}

	for _, d := range detectors {
		opts, err := d.Detect(extractPath)
		if err != nil {
			o.Logger.Warn("%s detector failed: %v", d.Name(), err)
			continue
		}
		report.Optimizations = append(report.Optimizations, opts...)
	}
}

// isIOSArtifact checks if the artifact is an iOS artifact
func (o *Orchestrator) isIOSArtifact(artifactType types.ArtifactType) bool {
	return artifactType == types.ArtifactTypeIPA ||
		artifactType == types.ArtifactTypeApp ||
		artifactType == types.ArtifactTypeXCArchive
}

// generateOptimizations creates optimization recommendations from analysis results
func (o *Orchestrator) generateOptimizations(report *types.Report) []types.Optimization {
	// Start with existing optimizations (from analyzers like strip-symbols, etc.)
	optimizations := report.Optimizations

	// Create duplicate categorizer for intelligent filtering
	categorizer := detector.NewDuplicateCategorizer()

	// Add duplicate file optimizations (with intelligent filtering)
	for _, dup := range report.Duplicates {
		// Evaluate duplicate with categorization rules
		filterResult := categorizer.EvaluateDuplicate(dup)

		// Skip if should be filtered out (architectural pattern or third-party SDK)
		if filterResult.ShouldFilter {
			// Optional: Log filtered duplicates for debugging
			// o.Logger.Debug("Filtered duplicate: %s (reason: %s)", dup.Files[0], filterResult.Reason)
			continue
		}

		// This is an actionable duplicate - create optimization
		severity := filterResult.Priority
		if severity == "" {
			// No priority specified by rule, calculate based on size
			severity = getSeverity(dup.WastedSize, report.ArtifactInfo.Size)
		}

		optimizations = append(optimizations, types.Optimization{
			Category:    "duplicates",
			Severity:    severity,
			Title:       fmt.Sprintf("Remove %d duplicate copies of files", dup.Count-1),
			Description: fmt.Sprintf("Found %d identical files (%s each)", dup.Count, util.FormatBytes(dup.Size)),
			Impact:      dup.WastedSize,
			Files:       dup.Files,
			Action:      "Keep only one copy and deduplicate references",
		})
	}

	return optimizations
}

// getSeverity determines severity based on impact relative to total size
func getSeverity(impact, totalSize int64) string {
	if totalSize == 0 {
		return "low"
	}

	percentage := float64(impact) / float64(totalSize) * 100

	if percentage >= 10 {
		return "high"
	} else if percentage >= 5 {
		return "medium"
	}
	return "low"
}

// calculateTotalSavings sums up all potential savings from optimizations
func calculateTotalSavings(report *types.Report) int64 {
	var total int64
	for _, opt := range report.Optimizations {
		total += opt.Impact
	}
	return total
}

// enrichWithCIMetadata adds Git/CI metadata from environment variables
func (o *Orchestrator) enrichWithCIMetadata(report *types.Report) {
	if report.Metadata == nil {
		report.Metadata = make(map[string]interface{})
	}

	// Try Bitrise environment variables first
	if branch := os.Getenv("BITRISE_GIT_BRANCH"); branch != "" {
		report.Metadata["git_branch"] = branch
	} else if branch := os.Getenv("GIT_BRANCH"); branch != "" {
		report.Metadata["git_branch"] = branch
	}

	// Use commit_hash key to match markdown formatter expectations
	// Try Bitrise environment variable first (GIT_CLONE_COMMIT_HASH)
	if commit := os.Getenv("GIT_CLONE_COMMIT_HASH"); commit != "" {
		report.Metadata["commit_hash"] = commit
	} else if commit := os.Getenv("BITRISE_GIT_COMMIT"); commit != "" {
		report.Metadata["commit_hash"] = commit
	} else if commit := os.Getenv("GIT_COMMIT"); commit != "" {
		report.Metadata["commit_hash"] = commit
	}
}

// detectAssetDuplicates extracts asset catalogs from report metadata and detects duplicates.
func (o *Orchestrator) detectAssetDuplicates(report *types.Report) []types.DuplicateSet {
	if report.Metadata == nil {
		return nil
	}

	// Extract asset catalogs from metadata
	catalogsInterface, ok := report.Metadata["asset_catalogs"]
	if !ok {
		return nil
	}

	// Convert to internal asset catalog type
	var internalCatalogs []*assets.AssetCatalogInfo

	switch catalogs := catalogsInterface.(type) {
	case []*types.AssetCatalogInfo:
		// Convert types.AssetCatalogInfo to assets.AssetCatalogInfo
		for _, tc := range catalogs {
			if tc == nil {
				continue
			}
			ic := &assets.AssetCatalogInfo{
				Path:       tc.Path,
				TotalSize:  tc.TotalSize,
				AssetCount: tc.AssetCount,
				ByType:     tc.ByType,
				ByScale:    tc.ByScale,
			}
			// Convert assets
			for _, ta := range tc.Assets {
				ic.Assets = append(ic.Assets, assets.AssetInfo{
					Name:          ta.Name,
					RenditionName: ta.RenditionName,
					Type:          ta.Type,
					Scale:         ta.Scale,
					Size:          ta.Size,
					Idiom:         ta.Idiom,
					Compression:   ta.Compression,
					PixelWidth:    ta.PixelWidth,
					PixelHeight:   ta.PixelHeight,
					SHA1Digest:    ta.SHA1Digest,
				})
			}
			internalCatalogs = append(internalCatalogs, ic)
		}
	case []*assets.AssetCatalogInfo:
		internalCatalogs = catalogs
	default:
		o.Logger.Warn("asset_catalogs metadata has unexpected type: %T", catalogsInterface)
		return nil
	}

	if len(internalCatalogs) == 0 {
		return nil
	}

	// Run asset duplicate detection
	assetDupDetector := detector.NewAssetDuplicateDetector()
	return assetDupDetector.DetectAssetDuplicates(internalCatalogs)
}
