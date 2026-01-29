package ios

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Create size breakdown
	sizeBreakdown := categorizeSizes(fileTree)

	// Find largest files
	largestFiles := findLargestFiles(fileTree, 10)

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:             path,
			Type:             types.ArtifactTypeApp,
			Size:             totalSize, // For .app, size is uncompressed
			UncompressedSize: totalSize,
			AnalyzedAt:       time.Now(),
		},
		SizeBreakdown: sizeBreakdown,
		FileTree:      fileTree,
		LargestFiles:  largestFiles,
		Metadata: map[string]interface{}{
			"app_bundle":   filepath.Base(path),
			"is_directory": true,
			"binaries":     binaries,
		},
	}

	return report, nil
}
