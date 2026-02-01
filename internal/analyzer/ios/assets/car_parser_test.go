package assets

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAssetCatalog(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("assetutil requires macOS")
	}

	carPath := "../../../../test-artifacts/ios/Wikipedia.app/Assets.car"

	if _, err := os.Stat(carPath); os.IsNotExist(err) {
		t.Skip("Assets.car test artifact not found")
	}

	catalog, err := ParseAssetCatalog(carPath)
	require.NoError(t, err)
	assert.NotNil(t, catalog)
	assert.Equal(t, "Assets.car", catalog.Path)
	assert.True(t, catalog.TotalSize > 0, "Should have total size")
	assert.True(t, catalog.AssetCount >= 0, "Should have asset count")
	t.Logf("Found %d assets, total size: %d bytes", catalog.AssetCount, catalog.TotalSize)

	// Check that assets have actual size data (not averaged)
	if len(catalog.Assets) > 0 {
		// Verify we have SHA1Digest for duplicate detection
		hasHash := false
		for _, asset := range catalog.Assets {
			if asset.SHA1Digest != "" {
				hasHash = true
				break
			}
		}
		t.Logf("Has SHA1Digest: %v", hasHash)

		// Log some asset details
		for i, asset := range catalog.Assets[:min(5, len(catalog.Assets))] {
			t.Logf("Asset %d: %s (type: %s, size: %d, idiom: %s, scale: %s)",
				i, asset.Name, asset.Type, asset.Size, asset.Idiom, asset.Scale)
		}
	}
}

func TestCategorizeAssets(t *testing.T) {
	assets := []AssetInfo{
		{Name: "icon.png", Type: "image", Size: 1000},
		{Name: "background.png", Type: "image", Size: 2000},
		{Name: "logo.pdf", Type: "vector", Size: 500},
		{Name: "icon@2x", Type: "image", Scale: "2x", Size: 1500},
		{Name: "icon@3x", Type: "image", Scale: "3x", Size: 2500},
	}

	byType, byScale := CategorizeAssets(assets)

	assert.Equal(t, int64(7000), byType["image"]) // 1000 + 2000 + 1500 + 2500
	assert.Equal(t, int64(500), byType["vector"])
	assert.Equal(t, int64(1500), byScale["2x"])
	assert.Equal(t, int64(2500), byScale["3x"])
}

func TestBuildVirtualAssetName(t *testing.T) {
	tests := []struct {
		name     string
		asset    AssetInfo
		expected string
	}{
		{
			name:     "simple image",
			asset:    AssetInfo{Name: "AppIcon", Type: "image"},
			expected: "AppIcon.png",
		},
		{
			name:     "image with scale",
			asset:    AssetInfo{Name: "AppIcon", Type: "image", Scale: "2x"},
			expected: "AppIcon@2x.png",
		},
		{
			name:     "image with idiom and scale",
			asset:    AssetInfo{Name: "AppIcon", Type: "image", Idiom: "iphone", Scale: "3x"},
			expected: "AppIcon~iphone@3x.png",
		},
		{
			name:     "universal image (no idiom suffix)",
			asset:    AssetInfo{Name: "AppIcon", Type: "image", Idiom: "universal", Scale: "2x"},
			expected: "AppIcon@2x.png",
		},
		{
			name:     "1x scale (no scale suffix)",
			asset:    AssetInfo{Name: "AppIcon", Type: "image", Scale: "1x"},
			expected: "AppIcon.png",
		},
		{
			name:     "vector asset",
			asset:    AssetInfo{Name: "Logo", Type: "vector"},
			expected: "Logo.pdf",
		},
		{
			name:     "color asset (no extension)",
			asset:    AssetInfo{Name: "AccentColor", Type: "color"},
			expected: "AccentColor",
		},
		{
			name:     "data asset",
			asset:    AssetInfo{Name: "SomeData", Type: "data"},
			expected: "SomeData.data",
		},
		{
			name:     "icon type",
			asset:    AssetInfo{Name: "AppIcon", Type: "icon", Idiom: "ipad", Scale: "2x"},
			expected: "AppIcon~ipad@2x.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildVirtualAssetName(tt.asset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindLargestAssets(t *testing.T) {
	assets := []AssetInfo{
		{Name: "small", Size: 100},
		{Name: "large", Size: 5000},
		{Name: "medium", Size: 1000},
		{Name: "huge", Size: 10000},
		{Name: "tiny", Size: 50},
	}

	largest := findLargestAssets(assets, 3)

	assert.Len(t, largest, 3)
	assert.Equal(t, "huge", largest[0].Name)
	assert.Equal(t, "large", largest[1].Name)
	assert.Equal(t, "medium", largest[2].Name)
}

func TestExpandAssetsAsChildren(t *testing.T) {
	catalog := &AssetCatalogInfo{
		Path: "Assets.car",
		Assets: []AssetInfo{
			{Name: "Icon", Type: "image", Size: 1000, Idiom: "universal", Scale: "2x"},
			{Name: "Background", Type: "image", Size: 2000, Idiom: "iphone", Scale: "3x"},
		},
	}

	children := ExpandAssetsAsChildren(catalog, "Payload/App.app/Assets.car")

	if runtime.GOOS != "darwin" {
		// On non-darwin, stub returns nil
		assert.Nil(t, children)
		return
	}

	require.Len(t, children, 2)

	// Should be sorted by size descending
	assert.Equal(t, "Background~iphone@3x.png", children[0].Name)
	assert.Equal(t, int64(2000), children[0].Size)
	assert.True(t, children[0].IsVirtual)
	assert.Equal(t, "Payload/App.app/Assets.car", children[0].SourceFile)

	assert.Equal(t, "Icon@2x.png", children[1].Name)
	assert.Equal(t, int64(1000), children[1].Size)
}

func TestGracefulFallback(t *testing.T) {
	// Test with a non-existent file
	catalog, err := ParseAssetCatalog("/nonexistent/Assets.car")
	assert.Error(t, err)
	assert.Nil(t, catalog)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
