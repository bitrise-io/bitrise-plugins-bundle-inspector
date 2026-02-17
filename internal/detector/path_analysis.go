package detector

import (
	"path/filepath"
	"regexp"
	"strings"
)

// localeDirectoryPattern matches locale directory segments like "es_419", "zh-CN_ALL", "he_ALL", "en_US".
// Pattern: 2-letter lowercase language code followed by one or more segments of separator (_/-)
// plus either an uppercase country/script code (2-4 chars) or a 3-digit numeric region code.
// This avoids false positives on generic directory names like "my_module" or "go_test".
var localeDirectoryPattern = regexp.MustCompile(`^[a-z]{2}(?:[_-](?:[A-Z]{2,4}|[0-9]{3}))+$`)

// PathAnalyzer provides utilities for analyzing iOS/Android artifact paths
type PathAnalyzer struct {
	rootPath string
}

// NewPathAnalyzer creates a new PathAnalyzer
func NewPathAnalyzer(rootPath string) *PathAnalyzer {
	return &PathAnalyzer{
		rootPath: rootPath,
	}
}

// BundleInfo contains information about a bundle boundary
type BundleInfo struct {
	Name     string // Bundle name (e.g., "GoogleMaps.framework")
	Type     string // Bundle type (e.g., "framework", "appex", "app", "xcassets", "lproj")
	InBundle bool   // Whether path is inside a bundle
	FullPath string // Full bundle path
}

// GetBundleBoundaries analyzes a path and returns bundle information
// Detects: .framework, .appex, .app, .xcassets, .lproj
// Returns the innermost (last) bundle in the path for nested bundles
func (p *PathAnalyzer) GetBundleBoundaries(path string) BundleInfo {
	// Normalize path separators
	path = filepath.ToSlash(path)

	// Priority order: check more specific bundles first
	// This ensures we get framework/appex before app, and xcassets/lproj before their parents
	bundleExtensions := []string{".framework", ".appex", ".xcassets", ".lproj", ".app"}

	var lastMatch BundleInfo
	lastMatch.InBundle = false

	// Find the last occurrence of any bundle type
	lastIdx := -1
	lastExt := ""

	for _, ext := range bundleExtensions {
		// Find the last occurrence of this extension
		idx := strings.LastIndex(path, ext+"/")
		if idx > lastIdx {
			lastIdx = idx
			lastExt = ext
		}
	}

	if lastIdx != -1 {
		// Found a bundle - extract its info
		start := strings.LastIndex(path[:lastIdx], "/")
		if start == -1 {
			start = 0
		} else {
			start++ // Skip the slash
		}

		bundleType := strings.TrimPrefix(lastExt, ".")
		bundlePath := path[:lastIdx+len(lastExt)]
		bundleName := path[start : lastIdx+len(lastExt)]

		lastMatch = BundleInfo{
			Name:     bundleName,
			Type:     bundleType,
			InBundle: true,
			FullPath: bundlePath,
		}
	}

	return lastMatch
}

// IsFrameworkPath checks if path is inside a .framework bundle
func (p *PathAnalyzer) IsFrameworkPath(path string) bool {
	path = filepath.ToSlash(path)
	return strings.Contains(path, ".framework/")
}

// IsExtensionPath checks if path is inside a .appex bundle (app extension)
func (p *PathAnalyzer) IsExtensionPath(path string) bool {
	path = filepath.ToSlash(path)
	return strings.Contains(path, ".appex/")
}

// IsAssetCatalogPath checks if path is inside an asset catalog (.xcassets or .car)
func (p *PathAnalyzer) IsAssetCatalogPath(path string) bool {
	path = filepath.ToSlash(path)
	return strings.Contains(path, ".xcassets/") || strings.Contains(path, ".car/") || strings.HasSuffix(path, ".car")
}

// IsLocalizationPath checks if path is inside a localization bundle (.lproj)
func (p *PathAnalyzer) IsLocalizationPath(path string) bool {
	path = filepath.ToSlash(path)
	return strings.Contains(path, ".lproj/")
}

// ExtractFrameworkName extracts the framework name from a framework path
// Example: "Payload/App.app/Frameworks/GoogleMaps.framework/..." -> "GoogleMaps"
func (p *PathAnalyzer) ExtractFrameworkName(path string) string {
	path = filepath.ToSlash(path)

	if idx := strings.Index(path, ".framework/"); idx != -1 {
		start := strings.LastIndex(path[:idx], "/")
		if start == -1 {
			start = 0
		} else {
			start++
		}
		return path[start:idx]
	}

	return ""
}

// ExtractExtensionName extracts the extension name from an extension path
// Example: "Payload/App.app/PlugIns/ShareExtension.appex/..." -> "ShareExtension"
func (p *PathAnalyzer) ExtractExtensionName(path string) string {
	path = filepath.ToSlash(path)

	if idx := strings.Index(path, ".appex/"); idx != -1 {
		start := strings.LastIndex(path[:idx], "/")
		if start == -1 {
			start = 0
		} else {
			start++
		}
		return path[start:idx]
	}

	return ""
}

// GetDistinctBundles returns a list of distinct bundle paths from a list of file paths
func (p *PathAnalyzer) GetDistinctBundles(paths []string) []string {
	bundles := make(map[string]bool)

	for _, path := range paths {
		info := p.GetBundleBoundaries(path)
		if info.InBundle {
			bundles[info.FullPath] = true
		}
	}

	result := make([]string, 0, len(bundles))
	for bundle := range bundles {
		result = append(result, bundle)
	}

	return result
}

// AreInDifferentBundles checks if two paths are in different bundles
// Returns true if they're in different bundles of the same type
func (p *PathAnalyzer) AreInDifferentBundles(path1, path2 string) bool {
	info1 := p.GetBundleBoundaries(path1)
	info2 := p.GetBundleBoundaries(path2)

	// Both must be in bundles
	if !info1.InBundle || !info2.InBundle {
		return false
	}

	// Must be same bundle type
	if info1.Type != info2.Type {
		return false
	}

	// Must be different bundles
	return info1.FullPath != info2.FullPath
}

// GetFileName extracts the filename from a path
func (p *PathAnalyzer) GetFileName(path string) string {
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// GetFileExtension extracts the file extension (without dot)
func (p *PathAnalyzer) GetFileExtension(path string) string {
	name := p.GetFileName(path)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		return name[idx+1:]
	}
	return ""
}

// ExtractLocaleDirectory finds a locale directory segment in the path and returns it.
// Matches patterns like "es_419", "zh-CN_ALL", "he_ALL", "en_US", "iw_ALL".
// Returns the locale segment and true if found, empty string and false otherwise.
func (p *PathAnalyzer) ExtractLocaleDirectory(path string) (string, bool) {
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if localeDirectoryPattern.MatchString(part) {
			return part, true
		}
	}

	return "", false
}

// ReplaceLocaleDirectory replaces the locale directory segment in a path with a placeholder.
// This allows comparing paths that differ only by locale. Returns the normalized path and
// whether a locale segment was found.
func (p *PathAnalyzer) ReplaceLocaleDirectory(path string) (string, bool) {
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if localeDirectoryPattern.MatchString(part) {
			parts[i] = "<LOCALE>"
			return strings.Join(parts, "/"), true
		}
	}

	return path, false
}
