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
			largestAssets[j] = convertAssetInfo(asset)
		}

		// Convert all assets
		allAssets := make([]types.AssetInfo, len(catalog.Assets))
		for j, asset := range catalog.Assets {
			allAssets[j] = convertAssetInfo(asset)
		}

		typedAssetCatalogs[i] = &types.AssetCatalogInfo{
			Path:          catalog.Path,
			TotalSize:     catalog.TotalSize,
			AssetCount:    catalog.AssetCount,
			ByType:        catalog.ByType,
			ByScale:       catalog.ByScale,
			LargestAssets: largestAssets,
			Assets:        allAssets,
		}
	}
	return typedAssetCatalogs
}

// convertAssetInfo converts internal AssetInfo to types.AssetInfo
func convertAssetInfo(asset assets.AssetInfo) types.AssetInfo {
	return types.AssetInfo{
		Name:          asset.Name,
		RenditionName: asset.RenditionName,
		Type:          asset.Type,
		Scale:         asset.Scale,
		Size:          asset.Size,
		Idiom:         asset.Idiom,
		Compression:   asset.Compression,
		PixelWidth:    asset.PixelWidth,
		PixelHeight:   asset.PixelHeight,
		SHA1Digest:    asset.SHA1Digest,
	}
}
