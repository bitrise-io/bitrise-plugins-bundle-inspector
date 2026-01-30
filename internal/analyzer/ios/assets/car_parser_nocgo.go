//go:build !cgo || !darwin
// +build !cgo !darwin

package assets

import (
	"fmt"
	"os"
	"path/filepath"
)

// ParseAssetCatalog extracts metadata from an Assets.car file.
// This is a stub implementation for non-CGo builds that returns basic file info only.
func ParseAssetCatalog(carPath string) (*AssetCatalogInfo, error) {
	// Get file size
	fileInfo, err := os.Stat(carPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat Assets.car: %w", err)
	}

	// Return basic catalog info without parsing (requires CGo)
	catalog := &AssetCatalogInfo{
		Path:       filepath.Base(carPath),
		TotalSize:  fileInfo.Size(),
		AssetCount: 0, // Cannot determine without CGo
		ByType:     make(map[string]int64),
		ByScale:    make(map[string]int64),
	}

	return catalog, nil
}

// ExtractAssetMetadata lists all assets in the catalog.
// This is a stub implementation for non-CGo builds.
func ExtractAssetMetadata(carPath string) ([]AssetInfo, error) {
	return nil, fmt.Errorf("asset metadata extraction not supported in this build (requires CGo and macOS)")
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
