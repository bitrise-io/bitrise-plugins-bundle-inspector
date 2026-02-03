package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ExtensionDuplicationRule detects resource duplication between app and extensions
// These are actionable optimizations - extensions often duplicate app resources unnecessarily
type ExtensionDuplicationRule struct {
	analyzer *PathAnalyzer
}

// NewExtensionDuplicationRule creates a new extension duplication detection rule
func NewExtensionDuplicationRule() *ExtensionDuplicationRule {
	return &ExtensionDuplicationRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *ExtensionDuplicationRule) ID() string {
	return "rule-8-extension-duplication"
}

// Name returns the rule name
func (r *ExtensionDuplicationRule) Name() string {
	return "App Extension Resource Duplication Detection"
}

// calculatePriority determines priority based on file size
// High: >500KB, Medium: 100-500KB, Low: <100KB
func calculatePriority(size int64) string {
	if size > 500*1024 { // > 500KB
		return "high"
	} else if size > 100*1024 { // 100-500KB
		return "medium"
	}
	return "low" // < 100KB
}

// Evaluate checks if duplicates are between app and extensions (actionable with priority)
func (r *ExtensionDuplicationRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must have at least 2 files
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Check if any files are in extensions
	hasExtension := false
	hasApp := false
	hasFramework := false

	for _, file := range dup.Files {
		if r.analyzer.IsExtensionPath(file) {
			hasExtension = true
		} else if r.analyzer.IsFrameworkPath(file) {
			hasFramework = true
		} else {
			// Assume files not in extension or framework are app files
			hasApp = true
		}
	}

	// If duplicates span app and extensions, this is actionable
	// Extensions often duplicate app resources (logos, images, strings)
	if hasExtension && (hasApp || hasFramework) {
		priority := calculatePriority(dup.Size)

		return FilterResult{
			ShouldFilter: false, // Don't filter - this is actionable
			Reason:       "Resource duplication between app and extension",
			RuleID:       r.ID(),
			Priority:     priority,
		}
	}

	// If duplicates are only within extensions (no app involvement), also actionable
	if hasExtension && !hasApp && !hasFramework {
		// Multiple extensions duplicating resources
		distinctBundles := r.analyzer.GetDistinctBundles(dup.Files)
		if len(distinctBundles) >= 2 {
			priority := calculatePriority(dup.Size)

			return FilterResult{
				ShouldFilter: false,
				Reason:       "Resource duplication between multiple extensions",
				RuleID:       r.ID(),
				Priority:     priority,
			}
		}
	}

	// Not an extension duplication pattern
	return FilterResult{ShouldFilter: false}
}
