package detector

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ContentsJSONRule detects Contents.json files in asset catalogs
// These are required metadata files for iOS asset catalogs
type ContentsJSONRule struct {
	analyzer *PathAnalyzer
}

// NewContentsJSONRule creates a new Contents.json detection rule
func NewContentsJSONRule() *ContentsJSONRule {
	return &ContentsJSONRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *ContentsJSONRule) ID() string {
	return "rule-3-contents-json"
}

// Name returns the rule name
func (r *ContentsJSONRule) Name() string {
	return "Asset Catalog Contents.json Detection"
}

// Evaluate checks if duplicate Contents.json files are in asset catalogs
func (r *ContentsJSONRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must be Contents.json files
	fileName := r.analyzer.GetFileName(dup.Files[0])
	if fileName != "Contents.json" {
		return FilterResult{ShouldFilter: false}
	}

	// Need at least 2 files to be duplicates
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Check if all files are in asset catalog paths (.xcassets or Assets.car)
	allInAssetCatalogs := true
	for _, file := range dup.Files {
		if !r.analyzer.IsAssetCatalogPath(file) {
			allInAssetCatalogs = false
			break
		}
	}

	// If all Contents.json files are in asset catalogs, this is legitimate metadata
	if allInAssetCatalogs {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Contents.json files in asset catalogs (required iOS metadata)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Check if files are in different asset sets
	// Example: AppIcon.appiconset/Contents.json vs LaunchImage.launchimage/Contents.json
	assetSets := make(map[string]bool)
	for _, file := range dup.Files {
		// Look for typical asset set directories
		for _, setType := range []string{".appiconset/", ".launchimage/", ".imageset/", ".colorset/", ".dataset/"} {
			if idx := strings.LastIndex(file, setType); idx != -1 {
				setDir := file[:idx+len(setType)-1] // Include set directory, exclude trailing slash
				assetSets[setDir] = true
				break
			}
		}
	}

	// If files are in different asset sets, this is legitimate
	if len(assetSets) >= 2 {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Contents.json files in different asset sets (required iOS metadata)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// If at least one file is in an asset catalog, likely legitimate
	// (Contents.json is always metadata in .xcassets context)
	hasAssetCatalogPath := false
	for _, file := range dup.Files {
		if strings.Contains(file, ".xcassets/") {
			hasAssetCatalogPath = true
			break
		}
	}

	if hasAssetCatalogPath {
		return FilterResult{
			ShouldFilter: true,
			Reason:       "Contents.json in asset catalog context (required iOS metadata)",
			RuleID:       r.ID(),
			Priority:     "",
		}
	}

	// Not in asset catalog context - potentially actionable
	return FilterResult{ShouldFilter: false}
}
