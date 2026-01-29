// Package analyzer provides common interfaces and utilities for artifact analysis.
package analyzer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/android"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// Analyzer defines the interface for artifact analyzers.
type Analyzer interface {
	// Analyze performs analysis on the artifact at the given path.
	Analyze(ctx context.Context, path string) (*types.Report, error)

	// ValidateArtifact checks if the file at path is a valid artifact of this type.
	ValidateArtifact(path string) error
}

// DetectArtifactType determines the artifact type from file extension.
func DetectArtifactType(path string) (types.ArtifactType, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".ipa":
		return types.ArtifactTypeIPA, nil
	case ".apk":
		return types.ArtifactTypeAPK, nil
	case ".aab":
		return types.ArtifactTypeAAB, nil
	case ".app":
		return types.ArtifactTypeApp, nil
	case ".xcarchive":
		return types.ArtifactTypeXCArchive, nil
	default:
		return "", fmt.Errorf("unsupported artifact type: %s", ext)
	}
}

// NewAnalyzer creates an appropriate analyzer for the given artifact path.
func NewAnalyzer(path string) (Analyzer, error) {
	artifactType, err := DetectArtifactType(path)
	if err != nil {
		return nil, err
	}

	switch artifactType {
	case types.ArtifactTypeIPA:
		return ios.NewIPAAnalyzer(), nil
	case types.ArtifactTypeAPK:
		return android.NewAPKAnalyzer(), nil
	case types.ArtifactTypeAAB:
		return android.NewAABAnalyzer(), nil
	case types.ArtifactTypeApp:
		return ios.NewAppAnalyzer(), nil
	case types.ArtifactTypeXCArchive:
		return nil, fmt.Errorf("XCArchive analysis not yet implemented")
	default:
		return nil, fmt.Errorf("no analyzer available for type: %s", artifactType)
	}
}
