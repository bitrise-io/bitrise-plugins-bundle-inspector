package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// InfoPlistRule detects Info.plist files in different bundles
// These are legitimate architectural patterns required by iOS
type InfoPlistRule struct {
	analyzer *PathAnalyzer
}

// NewInfoPlistRule creates a new Info.plist detection rule
func NewInfoPlistRule() *InfoPlistRule {
	return &InfoPlistRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *InfoPlistRule) ID() string {
	return "rule-1-info-plist"
}

// Name returns the rule name
func (r *InfoPlistRule) Name() string {
	return "Info.plist Bundle Boundary Detection"
}

// Evaluate checks if duplicate Info.plist files are in different bundles
func (r *InfoPlistRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must be Info.plist files
	if r.analyzer.GetFileName(dup.Files[0]) != "Info.plist" {
		return FilterResult{ShouldFilter: false}
	}

	// Need at least 2 files to be duplicates
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Check if files are in different bundles
	distinctBundles := r.analyzer.GetDistinctBundles(dup.Files)

	// If files are in different bundles, this is a legitimate pattern
	if len(distinctBundles) >= 2 {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Info.plist files in different bundles (required by iOS architecture)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Same bundle or no bundles - potentially actionable
	// (e.g., backup copies, duplicates in same directory)
	return FilterResult{ShouldFilter: false}
}
