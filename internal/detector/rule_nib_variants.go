package detector

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// NIBVariantsRule detects NIB version variants (runtime.nib vs objects-*.nib)
// These are legitimate iOS compatibility patterns
type NIBVariantsRule struct {
	analyzer *PathAnalyzer
}

// NewNIBVariantsRule creates a new NIB variants detection rule
func NewNIBVariantsRule() *NIBVariantsRule {
	return &NIBVariantsRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *NIBVariantsRule) ID() string {
	return "rule-2-nib-variants"
}

// Name returns the rule name
func (r *NIBVariantsRule) Name() string {
	return "NIB Version Variants Detection"
}

// Evaluate checks if duplicate NIB files are version variants
func (r *NIBVariantsRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must have at least 2 files
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Check if all files are NIB-related files
	allNIBFiles := true
	for _, file := range dup.Files {
		fileName := r.analyzer.GetFileName(file)
		if !strings.HasSuffix(fileName, ".nib") &&
			!strings.HasSuffix(fileName, "runtime.nib") &&
			!strings.HasPrefix(fileName, "objects-") {
			allNIBFiles = false
			break
		}
	}

	if !allNIBFiles {
		return FilterResult{ShouldFilter: false}
	}

	// Check for typical NIB variant patterns:
	// - runtime.nib and objects-*.nib in same .nib directory
	// - Multiple objects-*.nib files (different iOS versions)
	hasRuntimeNib := false
	hasObjectsNib := false

	for _, file := range dup.Files {
		fileName := r.analyzer.GetFileName(file)
		if fileName == "runtime.nib" || strings.HasSuffix(file, "/runtime.nib") {
			hasRuntimeNib = true
		}
		if strings.HasPrefix(fileName, "objects-") && strings.HasSuffix(fileName, ".nib") {
			hasObjectsNib = true
		}
	}

	// If we have both runtime.nib and objects-*.nib, this is a version variant pattern
	if hasRuntimeNib && hasObjectsNib {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "NIB version variants (runtime.nib vs objects-*.nib for iOS compatibility)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Check if files are in different .nib directories (different NIB bundles)
	// Example: Foo.nib/runtime.nib vs Bar.nib/runtime.nib
	nibDirs := make(map[string]bool)
	for _, file := range dup.Files {
		// Find the .nib directory
		if idx := strings.LastIndex(file, ".nib/"); idx != -1 {
			nibDir := file[:idx+4] // Include ".nib"
			nibDirs[nibDir] = true
		}
	}

	// If files are in different .nib directories, they're separate NIB bundles (legitimate)
	if len(nibDirs) >= 2 {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "NIB files in different .nib bundles (separate UI components)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Not a recognized NIB variant pattern
	return FilterResult{ShouldFilter: false}
}
