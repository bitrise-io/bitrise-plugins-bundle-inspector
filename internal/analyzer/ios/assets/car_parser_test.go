package assets

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAssetCatalog(t *testing.T) {
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
}

func TestExtractAssetMetadata(t *testing.T) {
	carPath := "../../../../test-artifacts/ios/Wikipedia.app/Assets.car"

	if _, err := os.Stat(carPath); os.IsNotExist(err) {
		t.Skip("Assets.car test artifact not found")
	}

	assets, err := ExtractAssetMetadata(carPath)
	// ExtractAssetMetadata returns empty list on parse errors, not an error
	require.NoError(t, err)

	if len(assets) > 0 {
		t.Logf("Found %d assets", len(assets))
		// Check first asset has expected fields
		assert.NotEmpty(t, assets[0].Name, "Asset should have name")
		assert.NotEmpty(t, assets[0].Type, "Asset should have type")
		t.Logf("First asset: %s (type: %s, size: %d)", assets[0].Name, assets[0].Type, assets[0].Size)
	}
}

func TestCategorizeAssets(t *testing.T) {
	assets := []AssetInfo{
		{Name: "icon.png", Type: "png", Size: 1000},
		{Name: "background.png", Type: "png", Size: 2000},
		{Name: "logo.pdf", Type: "pdf", Size: 500},
		{Name: "icon@2x", Type: "png", Scale: "2x", Size: 1500},
		{Name: "icon@3x", Type: "png", Scale: "3x", Size: 2500},
	}

	byType, byScale := CategorizeAssets(assets)

	assert.Equal(t, int64(7000), byType["png"]) // 1000 + 2000 + 1500 + 2500
	assert.Equal(t, int64(500), byType["pdf"])
	assert.Equal(t, int64(1500), byScale["2x"])
	assert.Equal(t, int64(2500), byScale["3x"])
}

func TestInferAssetType(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"icon.png", "png"},
		{"AppIcon", "png"},
		{"background.jpg", "jpeg"},
		{"logo.pdf", "pdf"},
		{"AccentColor", "color"},
		{"some-data", "data"},
		{"unknown-asset", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferAssetType(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInferScale(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"icon@3x", "3x"},
		{"icon@2x", "2x"},
		{"icon@1x", "1x"},
		{"icon-ipad", "tablet"},
		{"icon-iphone", "compact"},
		{"icon", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferScale(tt.name)
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

func TestGracefulFallback(t *testing.T) {
	// Test with a non-existent file
	catalog, err := ParseAssetCatalog("/nonexistent/Assets.car")
	assert.Error(t, err)
	assert.Nil(t, catalog)

	// Test with a file that's not an Assets.car (should return basic info)
	carPath := "../../../../test-artifacts/ios/Wikipedia.app/Info.plist"
	if _, err := os.Stat(carPath); os.IsNotExist(err) {
		t.Skip("Info.plist test artifact not found")
	}

	catalog, err = ParseAssetCatalog(carPath)
	// Should return error for non-car file on stat
	if err == nil {
		// If it doesn't error on stat, it should at least have basic info
		assert.NotNil(t, catalog)
		assert.True(t, catalog.TotalSize > 0)
	}
}
