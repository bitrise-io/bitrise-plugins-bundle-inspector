package detector

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// AssetDuplicationRule detects actual asset duplication within the same bundle
// These are true optimization opportunities - duplicated assets in the same context
type AssetDuplicationRule struct {
	analyzer *PathAnalyzer
}

// NewAssetDuplicationRule creates a new asset duplication detection rule
func NewAssetDuplicationRule() *AssetDuplicationRule {
	return &AssetDuplicationRule{
		analyzer: NewPathAnalyzer(""),
	}
}

// ID returns the rule identifier
func (r *AssetDuplicationRule) ID() string {
	return "rule-9-asset-duplication"
}

// Name returns the rule name
func (r *AssetDuplicationRule) Name() string {
	return "Asset Duplication in Same Bundle Detection"
}

// Common asset file extensions
var assetExtensions = map[string]bool{
	"png":  true,
	"jpg":  true,
	"jpeg": true,
	"gif":  true,
	"svg":  true,
	"pdf":  true, // PDF assets
	"mp3":  true,
	"m4a":  true,
	"wav":  true,
	"aac":  true,
	"mp4":  true,
	"mov":  true,
	"json": true, // JSON assets (animations, configs)
	"plist": true, // Configuration plists
}

// isAssetFile checks if a file is a common asset type
func (r *AssetDuplicationRule) isAssetFile(path string) bool {
	ext := r.analyzer.GetFileExtension(path)
	return assetExtensions[ext]
}

// Evaluate checks if duplicate assets are in the same bundle (actionable)
func (r *AssetDuplicationRule) Evaluate(dup types.DuplicateSet) FilterResult {
	// Must have at least 2 files
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// Check if all files are asset files
	allAssets := true
	for _, file := range dup.Files {
		if !r.isAssetFile(file) {
			allAssets = false
			break
		}
	}

	if !allAssets {
		return FilterResult{ShouldFilter: false}
	}

	// Check if files are in the same bundle
	distinctBundles := r.analyzer.GetDistinctBundles(dup.Files)

	// If all files are in the same bundle, this is a true duplication issue
	if len(distinctBundles) == 1 {
		priority := calculatePriority(dup.Size)

		return FilterResult{
			ShouldFilter: false, // Don't filter - this is actionable
			Reason:       "Asset duplication within same bundle (true duplicate)",
			RuleID:       r.ID(),
			Priority:     priority,
		}
	}

	// Check if files are in the same top-level directory (not in any bundle)
	// Example: Payload/App.app/image.png and Payload/App.app/Resources/image.png
	sameBundleContext := true
	var commonPrefix string

	for i, file := range dup.Files {
		// Get the first two path components (e.g., "Payload/App.app")
		parts := strings.Split(file, "/")
		if len(parts) >= 2 {
			prefix := strings.Join(parts[:2], "/")
			if i == 0 {
				commonPrefix = prefix
			} else if prefix != commonPrefix {
				sameBundleContext = false
				break
			}
		}
	}

	// If files share the same bundle context, this is actionable
	if sameBundleContext && commonPrefix != "" {
		priority := calculatePriority(dup.Size)

		return FilterResult{
			ShouldFilter: false,
			Reason:       "Asset duplication in same bundle context",
			RuleID:       r.ID(),
			Priority:     priority,
		}
	}

	// Not a same-bundle asset duplication pattern
	return FilterResult{ShouldFilter: false}
}
