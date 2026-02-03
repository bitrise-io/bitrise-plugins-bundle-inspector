package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// DuplicateCategorizer categorizes duplicate files into actionable vs. architectural patterns
type DuplicateCategorizer struct {
	registry *RuleRegistry
}

// NewDuplicateCategorizer creates a new duplicate categorizer with default rules
func NewDuplicateCategorizer() *DuplicateCategorizer {
	return &DuplicateCategorizer{
		registry: NewRuleRegistry(),
	}
}

// EvaluateDuplicate evaluates a single duplicate set against all rules
// Returns a FilterResult indicating if it should be filtered out or shown to user
func (c *DuplicateCategorizer) EvaluateDuplicate(dup types.DuplicateSet) FilterResult {
	return c.registry.Evaluate(dup)
}

// FilterDuplicates filters a list of duplicate sets, returning only actionable duplicates
// Returns:
// - actionable: duplicates that should be shown to user
// - filtered: duplicates that were filtered out (for debugging/logging)
func (c *DuplicateCategorizer) FilterDuplicates(duplicates []types.DuplicateSet) (actionable, filtered []types.DuplicateSet) {
	actionable = make([]types.DuplicateSet, 0)
	filtered = make([]types.DuplicateSet, 0)

	for _, dup := range duplicates {
		result := c.EvaluateDuplicate(dup)
		if result.ShouldFilter {
			filtered = append(filtered, dup)
		} else {
			actionable = append(actionable, dup)
		}
	}

	return actionable, filtered
}

// CategorizationResult contains the results of duplicate categorization
type CategorizationResult struct {
	Actionable []types.DuplicateSet // Duplicates that should be shown to user
	Filtered   []types.DuplicateSet // Duplicates that were filtered out
	TotalCount int                  // Total number of duplicate sets evaluated
}

// Categorize categorizes all duplicates and returns a structured result
func (c *DuplicateCategorizer) Categorize(duplicates []types.DuplicateSet) CategorizationResult {
	actionable, filtered := c.FilterDuplicates(duplicates)

	return CategorizationResult{
		Actionable: actionable,
		Filtered:   filtered,
		TotalCount: len(duplicates),
	}
}
