package detector

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// FrameworkMetadataRule detects framework metadata files
// These are required metadata files for framework distribution and usage
type FrameworkMetadataRule struct {
	analyzer *PathAnalyzer
}

// NewFrameworkMetadataRule creates a new framework metadata detection rule
func NewFrameworkMetadataRule() *FrameworkMetadataRule {
	return &FrameworkMetadataRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *FrameworkMetadataRule) ID() string {
	return "rule-6-framework-metadata"
}

// Name returns the rule name
func (r *FrameworkMetadataRule) Name() string {
	return "Framework Metadata Files Detection"
}

// Common framework metadata file extensions and names
var frameworkMetadataPatterns = []string{
	".supx",          // Carthage metadata
	".bcsymbolmap",   // Bitcode symbol maps
	".swiftdoc",      // Swift documentation (sometimes duplicated)
	".swiftmodule",   // Swift module files (sometimes duplicated)
	"module.modulemap", // Clang module maps
}

// Evaluate checks if duplicate files are framework metadata
func (r *FrameworkMetadataRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must have at least 2 files
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Check if files match metadata patterns
	fileName := r.analyzer.GetFileName(dup.Files[0])
	ext := r.analyzer.GetFileExtension(dup.Files[0])

	isMetadata := false
	for _, pattern := range frameworkMetadataPatterns {
		if strings.HasPrefix(pattern, ".") {
			// Extension pattern
			if "."+ext == pattern {
				isMetadata = true
				break
			}
		} else {
			// Filename pattern
			if fileName == pattern {
				isMetadata = true
				break
			}
		}
	}

	if !isMetadata {
		return FilterResult{ShouldFilter: false}
	}

	// Check if files are in framework paths
	allInFrameworks := true
	for _, file := range dup.Files {
		if !r.analyzer.IsFrameworkPath(file) {
			allInFrameworks = false
			break
		}
	}

	// If all metadata files are in frameworks, check if they're in different frameworks
	if allInFrameworks {
		distinctBundles := r.analyzer.GetDistinctBundles(dup.Files)
		if len(distinctBundles) >= 2 {
			return FilterResult{
				ShouldFilter: true,
				Reason:       "Framework metadata in different frameworks (required framework metadata)",
				RuleID:       r.ID(),
				Priority:     "",
			}
		}
	}

	// Special case: .bcsymbolmap files are often duplicated and should be filtered
	if ext == "bcsymbolmap" {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Bitcode symbol map files (build artifacts)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Not a recognized pattern
	return FilterResult{ShouldFilter: false}
}
