package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// maxSmallFileSize is the noise-reduction threshold (4KB) for duplicate detection.
// Duplicates at or below this size yield negligible savings on any platform:
// - iOS: each file occupies a full 4KB APFS block regardless of content size
// - Android: the absolute byte savings are too small to be actionable
const maxSmallFileSize = 4096

// SmallDuplicatesRule filters duplicate sets where the file size is at or below the
// noise-reduction threshold. These duplicates have negligible savings on any platform.
type SmallDuplicatesRule struct{}

// NewSmallDuplicatesRule creates a new small duplicates detection rule.
func NewSmallDuplicatesRule() *SmallDuplicatesRule {
	return &SmallDuplicatesRule{}
}

// ID returns the rule identifier.
func (r *SmallDuplicatesRule) ID() string {
	return "rule-10-small-duplicates"
}

// Name returns the rule name.
func (r *SmallDuplicatesRule) Name() string {
	return "Small File Duplicate Filter"
}

// Evaluate checks if the duplicate set consists of small files below the noise-reduction threshold.
func (r *SmallDuplicatesRule) Evaluate(dup types.DuplicateSet) FilterResult {
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	if dup.Size <= maxSmallFileSize {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Files at or below 4KB have negligible duplicate savings",
			RuleID:       r.ID(),
		}
	}

	return FilterResult{ShouldFilter: false}
}
