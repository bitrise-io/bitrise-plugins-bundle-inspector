package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// FilterResult indicates if a duplicate should be surfaced to the user
type FilterResult struct {
	ShouldFilter bool   // true = filter out (don't show), false = actionable (show)
	Reason       string // Human-readable reason (for debugging/logging)
	RuleID       string // Rule identifier (e.g., "rule-1-info-plist")
	Priority     string // "high", "medium", "low" (for actionable duplicates)
}

// Rule interface for duplicate detection rules
type Rule interface {
	ID() string                                  // Unique rule identifier
	Name() string                                // Human-readable rule name
	Evaluate(dup types.DuplicateSet) FilterResult // Evaluate a duplicate set
}

// RuleRegistry manages all duplicate detection rules
type RuleRegistry struct {
	rules []Rule
}

// NewRuleRegistry creates a new rule registry with default rules
func NewRuleRegistry() *RuleRegistry {
	registry := &RuleRegistry{
		rules: make([]Rule, 0),
	}

	// Register default rules (PR 1: Rules 1-3)
	registry.Register(NewInfoPlistRule())
	registry.Register(NewNIBVariantsRule())
	registry.Register(NewContentsJSONRule())

	// Register additional rules (PR 2: Rules 4-7)
	registry.Register(NewLocalizationRule())
	registry.Register(NewFrameworkScriptsRule())
	registry.Register(NewFrameworkMetadataRule())
	registry.Register(NewThirdPartySDKRule())

	// Register new filtering rules (before actionable rules)
	registry.Register(NewSmallDuplicatesRule())

	// Register actionable rules (PR 3: Rules 8-9)
	registry.Register(NewExtensionDuplicationRule())
	registry.Register(NewAssetDuplicationRule())

	return registry
}

// Register adds a rule to the registry
func (r *RuleRegistry) Register(rule Rule) {
	r.rules = append(r.rules, rule)
}

// Evaluate evaluates all rules against a duplicate set
// Returns the first matching rule's filter result, or a default "actionable" result if no rules match
func (r *RuleRegistry) Evaluate(dup types.DuplicateSet) FilterResult {
	for _, rule := range r.rules {
		result := rule.Evaluate(dup)
		if result.ShouldFilter || result.Priority != "" {
			// Rule matched and wants to filter or set priority
			return result
		}
	}

	// No rules matched - treat as actionable (show to user)
	return FilterResult{
		ShouldFilter: false,
		Reason:       "No filtering rules matched",
		RuleID:       "default",
		Priority:     "",
	}
}

// GetRules returns all registered rules
func (r *RuleRegistry) GetRules() []Rule {
	return r.rules
}
