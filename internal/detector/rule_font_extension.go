package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// fontExtensions are file extensions for font files.
var fontExtensions = map[string]bool{
	"ttf":   true,
	"otf":   true,
	"woff":  true,
	"woff2": true,
}

// FontExtensionRule filters font files duplicated between the main app/frameworks and extensions.
// iOS extensions run in separate sandboxed processes and cannot dynamically load fonts from the
// main app bundle or shared frameworks at runtime. Each extension must embed its own copy.
type FontExtensionRule struct {
	analyzer *PathAnalyzer
}

// NewFontExtensionRule creates a new font extension duplication rule.
func NewFontExtensionRule() *FontExtensionRule {
	return &FontExtensionRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier.
func (r *FontExtensionRule) ID() string {
	return "rule-12-font-extension"
}

// Name returns the rule name.
func (r *FontExtensionRule) Name() string {
	return "Font Extension Duplication Filter"
}

// Evaluate checks if duplicate font files span app/framework and extension boundaries.
func (r *FontExtensionRule) Evaluate(dup types.DuplicateSet) FilterResult {
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// All files must be font files
	for _, file := range dup.Files {
		ext := r.analyzer.GetFileExtension(file)
		if !fontExtensions[ext] {
			return FilterResult{ShouldFilter: false}
		}
	}

	// Check if duplication spans extension and app/framework
	hasExtension := false
	hasNonExtension := false
	for _, file := range dup.Files {
		if r.analyzer.IsExtensionPath(file) {
			hasExtension = true
		} else {
			hasNonExtension = true
		}
	}

	if hasExtension && hasNonExtension {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Font files required by sandboxed extensions (cannot load from main app bundle)",
			RuleID:       r.ID(),
		}
	}

	// Also filter fonts duplicated across multiple extensions (same sandboxing reason)
	if hasExtension {
		distinctBundles := r.analyzer.GetDistinctBundles(dup.Files)
		if len(distinctBundles) >= 2 {
			return FilterResult{
				ShouldFilter: true,
				Reason:       "Font files in different extensions (each sandbox requires its own copy)",
				RuleID:       r.ID(),
			}
		}
	}

	return FilterResult{ShouldFilter: false}
}
