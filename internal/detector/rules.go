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

// RuleConfig holds configuration options for rule registration.
type RuleConfig struct {
	// FilterSmallDuplicates controls whether files at or below filesystem block size (4KB)
	// are filtered out. When false, small duplicates are reported like any other duplicate.
	FilterSmallDuplicates bool
	// Platform controls which platform-specific rules are registered.
	// iOS-only rules (Info.plist, NIB, framework metadata, etc.) are skipped for Android.
	Platform Platform
}

// DefaultRuleConfig returns the default rule configuration.
func DefaultRuleConfig() RuleConfig {
	return RuleConfig{
		FilterSmallDuplicates: true,
	}
}

// RuleRegistry manages all duplicate detection rules
type RuleRegistry struct {
	rules []Rule
}

// NewRuleRegistry creates a new rule registry with default rules and default config.
func NewRuleRegistry() *RuleRegistry {
	return NewRuleRegistryWithConfig(DefaultRuleConfig())
}

// NewRuleRegistryWithConfig creates a new rule registry with the given configuration.
func NewRuleRegistryWithConfig(config RuleConfig) *RuleRegistry {
	registry := &RuleRegistry{
		rules: make([]Rule, 0),
	}

	// Register iOS-only filtering rules (Rules 1-7)
	// These reference iOS-specific concepts (Info.plist, NIB, .framework, .lproj, etc.)
	// that don't exist in Android bundles.
	if config.Platform != PlatformAndroid {
		registry.Register(NewInfoPlistRule())
		registry.Register(NewNIBVariantsRule())
		registry.Register(NewContentsJSONRule())
		registry.Register(NewLocalizationRule())
		registry.Register(NewFrameworkScriptsRule())
		registry.Register(NewFrameworkMetadataRule())
		registry.Register(NewThirdPartySDKRule())
	}

	// Register cross-platform filtering rules.
	// Rule 10 filters files <= 4KB as negligible savings.
	// Registered before actionable rules so small files never reach them.
	if config.FilterSmallDuplicates {
		registry.Register(NewSmallDuplicatesRule())
	}

	// Register iOS-only filtering rules (Rules 11-12) and actionable Rule 8.
	// These reference iOS-specific concepts (.appex extensions, device idioms).
	if config.Platform != PlatformAndroid {
		registry.Register(NewDeviceVariantRule())
		registry.Register(NewFontExtensionRule())
		registry.Register(NewExtensionDuplicationRule())
	}

	// Register cross-platform actionable rule (Rule 9)
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
