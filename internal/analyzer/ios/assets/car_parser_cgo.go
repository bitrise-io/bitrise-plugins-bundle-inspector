//go:build cgo && darwin
// +build cgo,darwin

package assets

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/iineva/bom/pkg/asset"
	"github.com/iineva/bom/pkg/bom"
)

// ParseAssetCatalog extracts metadata from an Assets.car file.
func ParseAssetCatalog(carPath string) (*AssetCatalogInfo, error) {
	// Get file size
	fileInfo, err := os.Stat(carPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat Assets.car: %w", err)
	}

	catalog := &AssetCatalogInfo{
		Path:       filepath.Base(carPath),
		TotalSize:  fileInfo.Size(),
		AssetCount: 0,
		ByType:     make(map[string]int64),
		ByScale:    make(map[string]int64),
	}

	// Try to parse the catalog
	assets, err := ExtractAssetMetadata(carPath)
	if err != nil {
		// Graceful fallback: return basic info from file stat
		return catalog, nil
	}

	// Update catalog with parsed data
	catalog.AssetCount = len(assets)

	// Categorize assets
	byType, byScale := CategorizeAssets(assets)
	catalog.ByType = byType
	catalog.ByScale = byScale

	// Find largest assets (top 10)
	catalog.LargestAssets = findLargestAssets(assets, 10)

	return catalog, nil
}

// ExtractAssetMetadata lists all assets in the catalog.
func ExtractAssetMetadata(carPath string) ([]AssetInfo, error) {
	f, err := os.Open(carPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Assets.car: %w", err)
	}
	defer f.Close()

	// Parse BOM structure
	b := bom.New(f)
	if err := b.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse BOM: %w", err)
	}

	// Try to create asset catalog reader
	f2, err := os.Open(carPath)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen Assets.car: %w", err)
	}
	defer f2.Close()

	assetCatalog, err := asset.NewWithReadSeeker(f2)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset catalog reader: %w", err)
	}

	var assets []AssetInfo

	// Read FACETKEYS tree to get asset names
	err = b.ReadTree("FACETKEYS", func(k io.Reader, d io.Reader) error {
		// Read the key (asset name) from the reader
		keyBytes, err := io.ReadAll(k)
		if err != nil {
			return nil // Skip on error
		}

		keyStr := string(keyBytes)

		// Skip empty or invalid keys
		if keyStr == "" {
			return nil
		}

		// Try to determine asset type and scale from name
		assetType := inferAssetType(keyStr)
		scale := inferScale(keyStr)

		// Estimate size (we can't easily get individual asset sizes from bom)
		// So we'll distribute the total size proportionally
		assets = append(assets, AssetInfo{
			Name:  keyStr,
			Type:  assetType,
			Scale: scale,
			Size:  0, // Will be estimated later
		})

		return nil
	})

	if err != nil {
		// If we can't read the tree, return empty list
		return []AssetInfo{}, nil
	}

	// Estimate individual asset sizes
	if len(assets) > 0 {
		fileInfo, _ := os.Stat(carPath)
		avgSize := fileInfo.Size() / int64(len(assets))
		for i := range assets {
			assets[i].Size = avgSize
		}
	}

	// Try to get more accurate sizes using the asset catalog reader
	// This is best effort - if it fails, we fall back to estimates
	for i := range assets {
		if img, err := assetCatalog.Image(assets[i].Name); err == nil && img != nil {
			// Image was found - we could potentially get more accurate info
			// For now, keep the estimate
			_ = img
		}
	}

	return assets, nil
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

// findLargestAssets returns the N largest assets.
func findLargestAssets(assets []AssetInfo, n int) []AssetInfo {
	if len(assets) == 0 {
		return nil
	}

	// Sort by size descending
	sorted := make([]AssetInfo, len(assets))
	copy(sorted, assets)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Size > sorted[j].Size
	})

	// Return top N
	if len(sorted) > n {
		sorted = sorted[:n]
	}

	return sorted
}

// inferAssetType tries to determine asset type from name.
func inferAssetType(name string) string {
	nameLower := strings.ToLower(name)

	if strings.Contains(nameLower, ".png") || strings.Contains(nameLower, "icon") || strings.Contains(nameLower, "image") {
		return "png"
	}
	if strings.Contains(nameLower, ".jpg") || strings.Contains(nameLower, ".jpeg") {
		return "jpeg"
	}
	if strings.Contains(nameLower, ".pdf") {
		return "pdf"
	}
	if strings.Contains(nameLower, ".svg") {
		return "svg"
	}
	if strings.Contains(nameLower, "color") {
		return "color"
	}
	if strings.Contains(nameLower, "data") {
		return "data"
	}

	return "unknown"
}

// inferScale tries to determine asset scale from name.
func inferScale(name string) string {
	if strings.Contains(name, "@3x") {
		return "3x"
	}
	if strings.Contains(name, "@2x") {
		return "2x"
	}
	if strings.Contains(name, "@1x") {
		return "1x"
	}

	// Check for other scale indicators
	if strings.Contains(name, "ipad") || strings.Contains(name, "tablet") {
		return "tablet"
	}
	if strings.Contains(name, "iphone") || strings.Contains(name, "compact") {
		return "compact"
	}

	return ""
}
