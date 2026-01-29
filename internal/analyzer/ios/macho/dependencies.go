package macho

import (
	"path/filepath"
	"strings"
)

// DependencyGraph represents the dependency relationships between binaries.
type DependencyGraph map[string][]string

// BuildDependencyGraph creates a dependency map from framework information.
// The map key is the framework/binary path, and the value is a list of dependencies.
func BuildDependencyGraph(frameworks map[string]*BinaryInfo) DependencyGraph {
	graph := make(DependencyGraph)

	for frameworkPath, binInfo := range frameworks {
		if binInfo == nil {
			continue
		}

		// Store this framework's dependencies
		deps := make([]string, 0)

		for _, lib := range binInfo.LinkedLibraries {
			// Skip system libraries
			if isSystemLibrary(lib) {
				continue
			}

			// Resolve @rpath and other special paths
			resolvedPath := resolveDependencyPath(lib, frameworkPath, binInfo.RPaths)
			if resolvedPath != "" {
				deps = append(deps, resolvedPath)
			}
		}

		graph[frameworkPath] = deps
	}

	return graph
}

// DetectUnusedFrameworks identifies frameworks that are not linked by the main binary
// or any other framework. Returns a list of potentially unused framework paths.
func DetectUnusedFrameworks(graph DependencyGraph, mainBinaryPath string) []string {
	// Build a set of all referenced frameworks
	referenced := make(map[string]bool)

	// Add direct dependencies of main binary
	if mainDeps, ok := graph[mainBinaryPath]; ok {
		for _, dep := range mainDeps {
			referenced[dep] = true
		}
	}

	// Recursively add transitive dependencies
	changed := true
	for changed {
		changed = false
		for fw := range referenced {
			if deps, ok := graph[fw]; ok {
				for _, dep := range deps {
					if !referenced[dep] {
						referenced[dep] = true
						changed = true
					}
				}
			}
		}
	}

	// Find frameworks that are not referenced
	var unused []string
	for frameworkPath := range graph {
		// Skip the main binary itself
		if frameworkPath == mainBinaryPath {
			continue
		}

		// Skip if referenced
		if referenced[frameworkPath] {
			continue
		}

		// This framework is not referenced
		unused = append(unused, frameworkPath)
	}

	return unused
}

// ResolveDependencyPath converts @rpath and other special path references to actual paths.
func resolveDependencyPath(libPath, binaryPath string, rpaths []string) string {
	// If it's a system library, return empty (we don't track these)
	if isSystemLibrary(libPath) {
		return ""
	}

	// Handle @rpath
	if strings.HasPrefix(libPath, "@rpath/") {
		// Remove @rpath/ prefix
		relPath := strings.TrimPrefix(libPath, "@rpath/")

		// Extract framework name from path like "Foo.framework/Foo"
		if strings.Contains(relPath, ".framework/") {
			frameworkName := strings.Split(relPath, ".framework/")[0] + ".framework"
			binaryName := strings.TrimSuffix(frameworkName, ".framework")
			return filepath.Join("Frameworks", frameworkName, binaryName)
		}
	}

	// Handle @executable_path
	if strings.HasPrefix(libPath, "@executable_path/") {
		relPath := strings.TrimPrefix(libPath, "@executable_path/")
		return relPath
	}

	// Handle @loader_path
	if strings.HasPrefix(libPath, "@loader_path/") {
		relPath := strings.TrimPrefix(libPath, "@loader_path/")
		// Relative to the loading binary's directory
		binDir := filepath.Dir(binaryPath)
		return filepath.Join(binDir, relPath)
	}

	// Absolute path or relative path
	return libPath
}

// resolveSpecialPath resolves @executable_path and @loader_path in rpaths.
func resolveSpecialPath(path, binaryPath string) string {
	if strings.HasPrefix(path, "@executable_path/") {
		return strings.TrimPrefix(path, "@executable_path/")
	}

	if strings.HasPrefix(path, "@loader_path/") {
		relPath := strings.TrimPrefix(path, "@loader_path/")
		binDir := filepath.Dir(binaryPath)
		return filepath.Join(binDir, relPath)
	}

	return path
}

// isSystemLibrary checks if a library path is a system library.
func isSystemLibrary(path string) bool {
	systemPrefixes := []string{
		"/System/Library/",
		"/usr/lib/",
		"/usr/local/lib/",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
