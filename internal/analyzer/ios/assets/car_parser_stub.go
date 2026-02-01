//go:build !darwin

package assets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ParseAssetCatalog extracts metadata from an Assets.car file.
// This is a stub implementation for non-macOS systems that returns basic file info only.
// Full asset parsing requires macOS assetutil command.
func ParseAssetCatalog(carPath string) (*AssetCatalogInfo, error) {
	// Get file size
	fileInfo, err := os.Stat(carPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat Assets.car: %w", err)
	}

	// Return basic catalog info without parsing (requires macOS)
	catalog := &AssetCatalogInfo{
		Path:       filepath.Base(carPath),
		TotalSize:  fileInfo.Size(),
		AssetCount: 0, // Cannot determine without assetutil
		ByType:     make(map[string]int64),
		ByScale:    make(map[string]int64),
	}

	return catalog, nil
}

// CategorizeAssets groups assets by type and scale.
func CategorizeAssets(assets []AssetInfo) (byType, byScale map[string]int64) {
	byType = make(map[string]int64)
	byScale = make(map[string]int64)

	for _, asset := range assets {
		byType[asset.Type] += asset.Size
		if asset.Scale != "" {
			byScale[asset.Scale] += asset.Size
		}
	}

	return byType, byScale
}

// BuildVirtualAssetName creates a filename for virtual asset display.
// This stub provides the same logic as the darwin version for consistency.
func BuildVirtualAssetName(asset AssetInfo) string {
	name := asset.Name

	// Add idiom if present
	if asset.Idiom != "" && asset.Idiom != "universal" {
		name += "~" + asset.Idiom
	}

	// Add scale if present and not 1x
	if asset.Scale != "" && asset.Scale != "1x" {
		name += "@" + asset.Scale
	}

	// Add extension based on type
	switch asset.Type {
	case "image", "icon", "imageset":
		name += ".png"
	case "vector":
		name += ".pdf"
	case "color":
		// No extension for colors
	case "data":
		name += ".data"
	default:
		if asset.Type != "" && asset.Type != "unknown" {
			name += "." + asset.Type
		}
	}

	return name
}

// ExpandAssetsAsChildren creates virtual FileNode children from asset catalog assets.
// On non-macOS systems, this returns nil as we cannot parse asset catalogs.
func ExpandAssetsAsChildren(catalog *AssetCatalogInfo, carRelativePath string) []*types.FileNode {
	// Cannot expand assets without assetutil (macOS only)
	return nil
}
