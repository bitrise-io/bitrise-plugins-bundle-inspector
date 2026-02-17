package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// maxSmallFileSize is the filesystem block size threshold (4KB).
// Files at or below this size occupy a single block on disk regardless of actual content size,
// making duplicate detection savings negligible or zero.
const maxSmallFileSize = 4096

// SmallDuplicatesRule filters duplicate sets where the file size is at or below the filesystem block size.
// These duplicates have negligible actual disk savings since each file occupies one block regardless.
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

// Evaluate checks if the duplicate set consists of small files below the filesystem block size.
func (r *SmallDuplicatesRule) Evaluate(dup types.DuplicateSet) FilterResult {
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	if dup.Size <= maxSmallFileSize {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Files at or below filesystem block size (4KB) have negligible duplicate savings",
			RuleID:       r.ID(),
		}
	}

	return FilterResult{ShouldFilter: false}
}
