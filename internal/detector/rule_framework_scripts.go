package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// FrameworkScriptsRule detects framework build scripts
// These are build-time artifacts that should be removed but are commonly found
type FrameworkScriptsRule struct {
	analyzer *PathAnalyzer
}

// NewFrameworkScriptsRule creates a new framework scripts detection rule
func NewFrameworkScriptsRule() *FrameworkScriptsRule {
	return &FrameworkScriptsRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *FrameworkScriptsRule) ID() string {
	return "rule-5-framework-scripts"
}

// Name returns the rule name
func (r *FrameworkScriptsRule) Name() string {
	return "Framework Build Scripts Detection"
}

// Common framework build script names
var frameworkScriptNames = []string{
	"strip-frameworks.sh",
	"copy-frameworks.sh",
	"embed-frameworks.sh",
	"run-script.sh",
	"strip-architectures.sh",
}

// Evaluate checks if duplicate files are framework build scripts
func (r *FrameworkScriptsRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must have at least 2 files
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Check if all files are known framework scripts
	fileName := r.analyzer.GetFileName(dup.Files[0])
	isFrameworkScript := false

	for _, scriptName := range frameworkScriptNames {
		if fileName == scriptName {
			isFrameworkScript = true
			break
		}
	}

	if !isFrameworkScript {
		return FilterResult{ShouldFilter: false}
	}

	// Check if files are in framework paths
	allInFrameworks := true
	for _, file := range dup.Files {
		if !r.analyzer.IsFrameworkPath(file) {
			allInFrameworks = false
			break
		}
	}

	// If all script files are in frameworks, this is a legitimate pattern
	// (These scripts are often bundled by dependency managers like CocoaPods)
	if allInFrameworks {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Framework build scripts in different frameworks (build artifacts, should be stripped)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Check if files are in different bundles
	distinctBundles := r.analyzer.GetDistinctBundles(dup.Files)
	if len(distinctBundles) >= 2 {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Build scripts in different bundles (build artifacts)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Not a recognized pattern
	return FilterResult{ShouldFilter: false}
}
