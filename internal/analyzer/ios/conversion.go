package ios

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ConvertToTypesBinaryInfo converts macho.BinaryInfo to types.BinaryInfo
func ConvertToTypesBinaryInfo(info *macho.BinaryInfo) *types.BinaryInfo {
	if info == nil {
		return nil
	}
	return &types.BinaryInfo{
		Architecture:     info.Architecture,
		Architectures:    info.Architectures,
		Type:             info.Type,
		CodeSize:         info.CodeSize,
		DataSize:         info.DataSize,
		LinkedLibraries:  info.LinkedLibraries,
		RPaths:           info.RPaths,
		HasDebugSymbols:  info.HasDebugSymbols,
		DebugSymbolsSize: info.DebugSymbolsSize,
	}
}

// ConvertToMachoBinaryInfo converts types.BinaryInfo to macho.BinaryInfo
func ConvertToMachoBinaryInfo(info *types.BinaryInfo) *macho.BinaryInfo {
	if info == nil {
		return nil
	}
	return &macho.BinaryInfo{
		Architecture:     info.Architecture,
		Architectures:    info.Architectures,
		Type:             info.Type,
		CodeSize:         info.CodeSize,
		DataSize:         info.DataSize,
		LinkedLibraries:  info.LinkedLibraries,
		RPaths:           info.RPaths,
		HasDebugSymbols:  info.HasDebugSymbols,
		DebugSymbolsSize: info.DebugSymbolsSize,
	}
}

// ConvertBinariesMapToMacho converts map of types.BinaryInfo to map of macho.BinaryInfo
func ConvertBinariesMapToMacho(binaries map[string]*types.BinaryInfo) map[string]*macho.BinaryInfo {
	result := make(map[string]*macho.BinaryInfo)
	for path, binInfo := range binaries {
		result[path] = ConvertToMachoBinaryInfo(binInfo)
	}
	return result
}

// ConvertFrameworksToTypes converts internal FrameworkInfo to types.FrameworkInfo
func ConvertFrameworksToTypes(frameworks []*FrameworkInfo) []*types.FrameworkInfo {
	typedFrameworks := make([]*types.FrameworkInfo, len(frameworks))
	for i, fw := range frameworks {
		binInfo := ConvertToTypesBinaryInfo(fw.BinaryInfo)
		typedFrameworks[i] = &types.FrameworkInfo{
			Name:         fw.Name,
			Path:         fw.Path,
			Version:      fw.Version,
			Size:         fw.Size,
			BinaryInfo:   binInfo,
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
