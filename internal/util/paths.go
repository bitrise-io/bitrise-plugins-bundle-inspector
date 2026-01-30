package util

import (
	"path/filepath"
	"strings"
)

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

// GetLowerExtension returns the lowercase file extension (e.g., ".png")
func GetLowerExtension(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

// HasExtension checks if a file has one of the specified extensions.
// Extensions should be provided in lowercase (e.g., ".png", ".jpg").
func HasExtension(path string, exts ...string) bool {
	ext := GetLowerExtension(path)
	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}
