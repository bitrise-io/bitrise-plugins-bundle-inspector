package detector

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// Detector is the interface for all optimization detectors
type Detector interface {
	// Detect runs the detector and returns optimization recommendations
	Detect(rootPath string) ([]types.Optimization, error)

	// Name returns the detector's name for logging
	Name() string
}

// PathMapper helps convert absolute paths to relative paths
type PathMapper struct {
	rootPath string
}

// NewPathMapper creates a new path mapper
func NewPathMapper(rootPath string) *PathMapper {
	return &PathMapper{rootPath: rootPath}
}

// ToRelative converts an absolute path to relative
func (m *PathMapper) ToRelative(absolutePath string) string {
	relPath := strings.TrimPrefix(absolutePath, m.rootPath)
	return strings.TrimPrefix(relPath, "/")
}

// ToRelativePaths converts multiple absolute paths to relative
func (m *PathMapper) ToRelativePaths(absolutePaths []string) []string {
	result := make([]string, len(absolutePaths))
	for i, path := range absolutePaths {
		result[i] = m.ToRelative(path)
	}
	return result
}
