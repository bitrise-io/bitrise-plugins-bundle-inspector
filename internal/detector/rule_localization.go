package detector

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// LocalizationRule detects localization files in different bundles
// These are legitimate patterns for bundle isolation (each bundle has its own localization)
type LocalizationRule struct {
	analyzer *PathAnalyzer
}

// NewLocalizationRule creates a new localization detection rule
func NewLocalizationRule() *LocalizationRule {
	return &LocalizationRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *LocalizationRule) ID() string {
	return "rule-4-localization"
}

// Name returns the rule name
func (r *LocalizationRule) Name() string {
	return "Localization File Bundle Isolation Detection"
}

// Evaluate checks if duplicate localization files are in different bundles
func (r *LocalizationRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must have at least 2 files
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// First: check locale-variant directory patterns (any file type)
	// These are files in different locale directories (e.g., es_419/, zh-CN_ALL/, he_ALL/)
	// that share identical content - applies to any file type, not just .strings
	if result := r.evaluateLocaleDirectories(dup); result.ShouldFilter {
		return result
	}

	// Below checks are for .strings/.stringsdict files only (iOS .lproj localization)
	allLocalizationFiles := true
	for _, file := range dup.Files {
		ext := r.analyzer.GetFileExtension(file)
		if ext != "strings" && ext != "stringsdict" {
			allLocalizationFiles = false
			break
		}
	}

	if !allLocalizationFiles {
		return FilterResult{ShouldFilter: false}
	}

	// Check if files are in different .lproj bundles
	lprojBundles := make(map[string]bool)
	allInLproj := true

	for _, file := range dup.Files {
		// Find the .lproj bundle
		if idx := strings.LastIndex(file, ".lproj/"); idx != -1 {
			lprojPath := file[:idx+6] // Include ".lproj"
			lprojBundles[lprojPath] = true
		} else {
			allInLproj = false
		}
	}

	// If all files are in different .lproj bundles, this is legitimate
	if allInLproj && len(lprojBundles) >= 2 {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Localization files in different .lproj bundles (bundle isolation pattern)",
			RuleID:       r.ID(),
		}
	}

	// Check if files are in different framework/extension bundles
	// Example: SDK.framework/en.lproj/Localizable.strings vs App.app/en.lproj/Localizable.strings
	distinctBundles := r.analyzer.GetDistinctBundles(dup.Files)

	// If files are in different bundles (framework, app, extension), this is legitimate
	// Each bundle should have its own localization
	if len(distinctBundles) >= 2 {
		// Additional check: ensure they're in the same locale
		// Example: both in en.lproj, de.lproj, etc.
		locales := make(map[string]bool)
		for _, file := range dup.Files {
			if idx := strings.LastIndex(file, ".lproj/"); idx != -1 {
				start := strings.LastIndex(file[:idx], "/")
				if start == -1 {
					start = 0
				} else {
					start++
				}
				locale := file[start : idx+6] // e.g., "en.lproj"
				locales[locale] = true
			}
		}

		// If same locale but different bundles, this is legitimate bundle isolation
		if len(locales) == 1 {
			return FilterResult{
				ShouldFilter: true,
				Reason:       "Localization files in different bundles with same locale (bundle isolation)",
				RuleID:       r.ID(),
			}
		}
	}

	return FilterResult{ShouldFilter: false}
}

// evaluateLocaleDirectories checks if files differ only by locale directory segment.
// Handles patterns like es_419/, es_AR/, es_MX/ for Spanish variants, or he_ALL/ vs iw_ALL/
// for legacy language code aliases.
func (r *LocalizationRule) evaluateLocaleDirectories(dup types.DuplicateSet) FilterResult {
	// All files must contain a locale directory segment
	normalizedPaths := make(map[string]bool)
	locales := make(map[string]bool)
	allHaveLocale := true

	for _, file := range dup.Files {
		normalized, found := r.analyzer.ReplaceLocaleDirectory(file)
		if !found {
			allHaveLocale = false
			break
		}

		locale, _ := r.analyzer.ExtractLocaleDirectory(file)
		normalizedPaths[normalized] = true
		locales[locale] = true
	}

	if !allHaveLocale {
		return FilterResult{ShouldFilter: false}
	}

	// If all files have the same path structure (only locale differs) and there are 2+ locales
	if len(normalizedPaths) == 1 && len(locales) >= 2 {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Files in different locale variant directories (language/region variants with identical content)",
			RuleID:       r.ID(),
		}
	}

	return FilterResult{ShouldFilter: false}
}
