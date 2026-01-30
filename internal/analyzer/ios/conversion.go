package ios

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ConvertFrameworksToTypes converts internal FrameworkInfo to types.FrameworkInfo
func ConvertFrameworksToTypes(frameworks []*FrameworkInfo) []*types.FrameworkInfo {
	typedFrameworks := make([]*types.FrameworkInfo, len(frameworks))
	for i, fw := range frameworks {
		typedFrameworks[i] = &types.FrameworkInfo{
			Name:         fw.Name,
			Path:         fw.Path,
			Version:      fw.Version,
			Size:         fw.Size,
			BinaryInfo:   fw.BinaryInfo, // No conversion needed - both use types.BinaryInfo
			Dependencies: fw.Dependencies,
		}
	}
	return typedFrameworks
}

// ConvertAssetCatalogsToTypes converts internal AssetCatalogInfo to types.AssetCatalogInfo
func ConvertAssetCatalogsToTypes(assetCatalogs []*assets.AssetCatalogInfo) []*types.AssetCatalogInfo {
	typedAssetCatalogs := make([]*types.AssetCatalogInfo, len(assetCatalogs))
	for i, catalog := range assetCatalogs {
		largestAssets := make([]types.AssetInfo, len(catalog.LargestAssets))
		for j, asset := range catalog.LargestAssets {
			largestAssets[j] = types.AssetInfo{
				Name:  asset.Name,
				Type:  asset.Type,
				Scale: asset.Scale,
				Size:  asset.Size,
			}
		}
		typedAssetCatalogs[i] = &types.AssetCatalogInfo{
			Path:          catalog.Path,
			TotalSize:     catalog.TotalSize,
			AssetCount:    catalog.AssetCount,
			ByType:        catalog.ByType,
			ByScale:       catalog.ByScale,
			LargestAssets: largestAssets,
		}
	}
	return typedAssetCatalogs
}
