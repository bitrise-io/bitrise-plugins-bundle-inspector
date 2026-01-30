package util

import "strings"

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
