package detector

import (
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// deviceIdiomSuffixes are the known iOS device idiom suffixes used in asset catalog virtual paths.
// These follow Apple's naming convention: ~<idiom> where idiom identifies the device class.
var deviceIdiomSuffixes = []string{
	"~iphone",
	"~ipad",
	"~phone",
	"~pad",
}

// scaleSuffixes are the known iOS asset scale factor suffixes.
var scaleSuffixes = []string{"@2x", "@3x", "@1x"}

// DeviceVariantRule filters duplicate sets that are device-specific variants of the same asset.
// iOS asset catalogs compile images for different device idioms (phone/pad) which may have
// identical content but are selected at runtime based on device type. Removing either copy
// breaks the app on that device class.
type DeviceVariantRule struct{}

// NewDeviceVariantRule creates a new device variant detection rule.
func NewDeviceVariantRule() *DeviceVariantRule {
	return &DeviceVariantRule{}
}

// ID returns the rule identifier.
func (r *DeviceVariantRule) ID() string {
	return "rule-11-device-variant"
}

// Name returns the rule name.
func (r *DeviceVariantRule) Name() string {
	return "Asset Catalog Device Variant Detection"
}

// Evaluate checks if duplicate files are device-specific variants of the same asset in a .car file.
func (r *DeviceVariantRule) Evaluate(dup types.DuplicateSet) FilterResult {
	if len(dup.Files) < 2 {
		return FilterResult{ShouldFilter: false}
	}

	// All files must be in .car directories
	for _, file := range dup.Files {
		if !strings.Contains(file, ".car/") {
			return FilterResult{ShouldFilter: false}
		}
	}

	// Extract base names (strip idiom suffix) and .car paths
	type parsed struct {
		carDir   string
		baseName string
		hasIdiom bool
	}

	entries := make([]parsed, len(dup.Files))
	anyHasIdiom := false

	for i, file := range dup.Files {
		dir := filepath.Dir(file)
		base := filepath.Base(file)

		// Strip extension
		ext := filepath.Ext(base)
		nameWithoutExt := strings.TrimSuffix(base, ext)

		// Strip scale suffix (@2x, @3x)
		for _, s := range scaleSuffixes {
			nameWithoutExt = strings.TrimSuffix(nameWithoutExt, s)
		}

		// Strip idiom suffix and check if present
		hasIdiom := false
		for _, idiom := range deviceIdiomSuffixes {
			if strings.HasSuffix(nameWithoutExt, idiom) {
				nameWithoutExt = strings.TrimSuffix(nameWithoutExt, idiom)
				hasIdiom = true
				break
			}
		}

		if hasIdiom {
			anyHasIdiom = true
		}

		entries[i] = parsed{
			carDir:   dir,
			baseName: nameWithoutExt,
			hasIdiom: hasIdiom,
		}
	}

	// At least one file must have an idiom suffix
	if !anyHasIdiom {
		return FilterResult{ShouldFilter: false}
	}

	// All must share the same .car directory and base name
	firstCar := entries[0].carDir
	firstBase := entries[0].baseName
	for _, e := range entries[1:] {
		if e.carDir != firstCar || e.baseName != firstBase {
			return FilterResult{ShouldFilter: false}
		}
	}

	return FilterResult{
		ShouldFilter: true,
		Reason:       "Asset catalog device variants (phone/pad idioms selected at runtime)",
		RuleID:       r.ID(),
	}
}
