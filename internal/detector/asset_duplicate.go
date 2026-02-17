package detector

import (
	"path/filepath"
	"sort"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// MinAssetDuplicateSavings is the minimum size (in bytes) for reporting asset duplicates.
// Assets smaller than this are not worth reporting (similar to Emerge Tools at ~0.5KB).
const MinAssetDuplicateSavings = 512

// AssetDuplicateDetector finds duplicate assets across .car files using SHA1Digest.
type AssetDuplicateDetector struct{}

// NewAssetDuplicateDetector creates a new asset duplicate detector.
func NewAssetDuplicateDetector() *AssetDuplicateDetector {
	return &AssetDuplicateDetector{}
}

// assetWithPath holds an asset and its full virtual path.
type assetWithPath struct {
	Asset       assets.AssetInfo
	VirtualPath string
}

// DetectAssetDuplicates finds duplicate assets across multiple asset catalogs.
// It groups assets by their SHA1Digest and returns DuplicateSets for groups with 2+ assets.
func (d *AssetDuplicateDetector) DetectAssetDuplicates(catalogs []*assets.AssetCatalogInfo) []types.DuplicateSet {
	// Group assets by SHA1Digest
	hashGroups := make(map[string][]assetWithPath)

	for _, catalog := range catalogs {
		if catalog == nil {
			continue
		}

		for _, asset := range catalog.Assets {
			// Skip assets without a hash
			if asset.SHA1Digest == "" {
				continue
			}

			virtualPath := buildFullVirtualPath(catalog.Path, asset)
			hashGroups[asset.SHA1Digest] = append(hashGroups[asset.SHA1Digest], assetWithPath{
				Asset:       asset,
				VirtualPath: virtualPath,
			})
		}
	}

	// Build DuplicateSet for groups with 2+ assets
	var duplicates []types.DuplicateSet
	for hash, assetGroup := range hashGroups {
		if len(assetGroup) < 2 {
			continue
		}

		size := assetGroup[0].Asset.Size

		// Skip if savings are too small (like Emerge Tools at ~0.5KB)
		if size < MinAssetDuplicateSavings {
			continue
		}

		// Skip device variant groups: same asset name in same .car file with different idioms
		// (e.g., Icon~phone.png and Icon~pad.png are runtime-selected, not true duplicates)
		if isDeviceVariantGroup(assetGroup) {
			continue
		}

		paths := make([]string, len(assetGroup))
		for i, a := range assetGroup {
			paths[i] = a.VirtualPath
		}

		wastedSize := (int64(len(assetGroup)) - 1) * size
		duplicates = append(duplicates, types.DuplicateSet{
			Hash:       hash,
			Size:       size,
			Count:      len(assetGroup),
			Files:      paths,
			WastedSize: wastedSize,
		})
	}

	// Sort by wasted size descending
	sort.Slice(duplicates, func(i, j int) bool {
		return duplicates[i].WastedSize > duplicates[j].WastedSize
	})

	return duplicates
}

// isDeviceVariantGroup checks if all assets in a hash group are device variants of the same asset.
// Device variants share the same Name and .car file but have different Idiom values
// (e.g., "iphone"/"ipad", "phone"/"pad"). These are intentionally created by Xcode's
// asset catalog compiler and selected at runtime based on device type.
func isDeviceVariantGroup(group []assetWithPath) bool {
	if len(group) < 2 {
		return false
	}

	firstName := group[0].Asset.Name
	firstCar := filepath.Dir(group[0].VirtualPath)

	// All assets must have the same name and be in the same .car file
	for _, a := range group[1:] {
		if a.Asset.Name != firstName {
			return false
		}
		if filepath.Dir(a.VirtualPath) != firstCar {
			return false
		}
	}

	// At least two different idioms must be present (otherwise it's a true duplicate)
	idioms := make(map[string]bool)
	for _, a := range group {
		idiom := a.Asset.Idiom
		if idiom == "" {
			idiom = "universal"
		}
		idioms[idiom] = true
	}

	return len(idioms) >= 2
}

// buildFullVirtualPath creates the full virtual path for an asset.
// Format: <car_relative_path>/<Name>~<idiom>@<scale>x.<extension>
func buildFullVirtualPath(carPath string, asset assets.AssetInfo) string {
	virtualName := assets.BuildVirtualAssetName(asset)
	return filepath.Join(carPath, virtualName)
}
