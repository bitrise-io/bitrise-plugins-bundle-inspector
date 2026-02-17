package detector

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/stretchr/testify/assert"
)

func TestAssetDuplicateDetector_DetectAssetDuplicates(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	catalogs := []*assets.AssetCatalogInfo{
		{
			Path: "Assets.car",
			Assets: []assets.AssetInfo{
				{Name: "Icon", Type: "image", Size: 1000, SHA1Digest: "abc123", Idiom: "universal", Scale: "2x"},
				{Name: "Logo", Type: "image", Size: 2000, SHA1Digest: "def456", Idiom: "iphone", Scale: "3x"},
			},
		},
		{
			Path: "Frameworks/MyFramework.framework/Assets.car",
			Assets: []assets.AssetInfo{
				// Duplicate of Icon
				{Name: "FrameworkIcon", Type: "image", Size: 1000, SHA1Digest: "abc123", Idiom: "universal", Scale: "2x"},
				{Name: "Unique", Type: "image", Size: 3000, SHA1Digest: "ghi789"},
			},
		},
	}

	duplicates := detector.DetectAssetDuplicates(catalogs)

	assert.Len(t, duplicates, 1)
	assert.Equal(t, "abc123", duplicates[0].Hash)
	assert.Equal(t, int64(1000), duplicates[0].Size)
	assert.Equal(t, 2, duplicates[0].Count)
	assert.Equal(t, int64(1000), duplicates[0].WastedSize) // (2-1) * 1000
	assert.Contains(t, duplicates[0].Files, "Assets.car/Icon@2x.png")
	assert.Contains(t, duplicates[0].Files, "Frameworks/MyFramework.framework/Assets.car/FrameworkIcon@2x.png")
}

func TestAssetDuplicateDetector_SkipsSmallDuplicates(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	catalogs := []*assets.AssetCatalogInfo{
		{
			Path: "Assets.car",
			Assets: []assets.AssetInfo{
				// Small duplicates (< 512 bytes) should be skipped
				{Name: "TinyIcon", Type: "image", Size: 100, SHA1Digest: "tiny123"},
			},
		},
		{
			Path: "Other.car",
			Assets: []assets.AssetInfo{
				{Name: "TinyIcon2", Type: "image", Size: 100, SHA1Digest: "tiny123"},
			},
		},
	}

	duplicates := detector.DetectAssetDuplicates(catalogs)

	assert.Len(t, duplicates, 0, "Small duplicates should be skipped")
}

func TestAssetDuplicateDetector_SkipsAssetsWithoutHash(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	catalogs := []*assets.AssetCatalogInfo{
		{
			Path: "Assets.car",
			Assets: []assets.AssetInfo{
				{Name: "Icon", Type: "image", Size: 1000, SHA1Digest: ""}, // No hash
			},
		},
		{
			Path: "Other.car",
			Assets: []assets.AssetInfo{
				{Name: "Icon2", Type: "image", Size: 1000, SHA1Digest: ""}, // No hash
			},
		},
	}

	duplicates := detector.DetectAssetDuplicates(catalogs)

	assert.Len(t, duplicates, 0, "Assets without SHA1Digest should be skipped")
}

func TestAssetDuplicateDetector_NilCatalogs(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	// Test with nil slice
	duplicates := detector.DetectAssetDuplicates(nil)
	assert.Nil(t, duplicates)

	// Test with empty slice
	duplicates = detector.DetectAssetDuplicates([]*assets.AssetCatalogInfo{})
	assert.Len(t, duplicates, 0)

	// Test with nil catalogs in slice
	catalogs := []*assets.AssetCatalogInfo{nil, nil}
	duplicates = detector.DetectAssetDuplicates(catalogs)
	assert.Len(t, duplicates, 0)
}

func TestAssetDuplicateDetector_MultipleDuplicates(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	catalogs := []*assets.AssetCatalogInfo{
		{
			Path: "Assets.car",
			Assets: []assets.AssetInfo{
				{Name: "Icon", Type: "icon", Size: 5000, SHA1Digest: "hash1", Idiom: "phone", Scale: "3x"},
				{Name: "Logo", Type: "image", Size: 3000, SHA1Digest: "hash2"},
			},
		},
		{
			Path: "Other.car",
			Assets: []assets.AssetInfo{
				{Name: "Icon2", Type: "icon", Size: 5000, SHA1Digest: "hash1", Idiom: "phone", Scale: "3x"},
				{Name: "Logo2", Type: "image", Size: 3000, SHA1Digest: "hash2"},
			},
		},
		{
			Path: "Third.car",
			Assets: []assets.AssetInfo{
				{Name: "Icon3", Type: "icon", Size: 5000, SHA1Digest: "hash1", Idiom: "phone", Scale: "3x"},
			},
		},
	}

	duplicates := detector.DetectAssetDuplicates(catalogs)

	// Should be sorted by wasted size descending
	assert.Len(t, duplicates, 2)
	assert.Equal(t, "hash1", duplicates[0].Hash) // 3 copies * 5000 = more waste
	assert.Equal(t, 3, duplicates[0].Count)
	assert.Equal(t, int64(10000), duplicates[0].WastedSize) // (3-1) * 5000

	assert.Equal(t, "hash2", duplicates[1].Hash) // 2 copies * 3000 = less waste
	assert.Equal(t, 2, duplicates[1].Count)
	assert.Equal(t, int64(3000), duplicates[1].WastedSize) // (2-1) * 3000
}

func TestAssetDuplicateDetector_SkipsDeviceVariants(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	catalogs := []*assets.AssetCatalogInfo{
		{
			Path: "Assets.car",
			Assets: []assets.AssetInfo{
				// Same asset, same name, same .car, different idiom - device variants
				{Name: "AppIcon", Type: "icon", Size: 5000, SHA1Digest: "variant_hash", Idiom: "phone", Scale: "2x"},
				{Name: "AppIcon", Type: "icon", Size: 5000, SHA1Digest: "variant_hash", Idiom: "pad", Scale: "2x"},
			},
		},
	}

	duplicates := detector.DetectAssetDuplicates(catalogs)
	assert.Len(t, duplicates, 0, "Device variants should not be reported as duplicates")
}

func TestAssetDuplicateDetector_KeepsCrossCatalogDuplicates(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	catalogs := []*assets.AssetCatalogInfo{
		{
			Path: "Assets.car",
			Assets: []assets.AssetInfo{
				{Name: "Icon", Type: "icon", Size: 5000, SHA1Digest: "cross_hash", Idiom: "phone", Scale: "2x"},
			},
		},
		{
			Path: "Frameworks/SDK.framework/Assets.car",
			Assets: []assets.AssetInfo{
				// Same hash but different .car - this IS a true duplicate
				{Name: "SDKIcon", Type: "icon", Size: 5000, SHA1Digest: "cross_hash", Idiom: "phone", Scale: "2x"},
			},
		},
	}

	duplicates := detector.DetectAssetDuplicates(catalogs)
	assert.Len(t, duplicates, 1, "Cross-catalog duplicates should still be reported")
}

func TestAssetDuplicateDetector_KeepsSameIdiomDuplicates(t *testing.T) {
	detector := NewAssetDuplicateDetector()

	catalogs := []*assets.AssetCatalogInfo{
		{
			Path: "Assets.car",
			Assets: []assets.AssetInfo{
				// Same name, same idiom, same .car - true duplicate (different rendition names)
				{Name: "Logo", Type: "image", Size: 3000, SHA1Digest: "same_hash", Idiom: "phone", Scale: "2x"},
				{Name: "Logo", Type: "image", Size: 3000, SHA1Digest: "same_hash", Idiom: "phone", Scale: "3x"},
			},
		},
	}

	duplicates := detector.DetectAssetDuplicates(catalogs)
	// Same idiom for both - NOT a device variant group, should be reported
	assert.Len(t, duplicates, 1, "Same-idiom duplicates should still be reported")
}

func TestBuildFullVirtualPath(t *testing.T) {
	tests := []struct {
		name     string
		carPath  string
		asset    assets.AssetInfo
		expected string
	}{
		{
			name:     "simple",
			carPath:  "Assets.car",
			asset:    assets.AssetInfo{Name: "Icon", Type: "image", Scale: "2x"},
			expected: "Assets.car/Icon@2x.png",
		},
		{
			name:     "nested car path",
			carPath:  "Frameworks/MyFramework.framework/Assets.car",
			asset:    assets.AssetInfo{Name: "Logo", Type: "vector"},
			expected: "Frameworks/MyFramework.framework/Assets.car/Logo.pdf",
		},
		{
			name:     "with idiom",
			carPath:  "Assets.car",
			asset:    assets.AssetInfo{Name: "Icon", Type: "icon", Idiom: "ipad", Scale: "2x"},
			expected: "Assets.car/Icon~ipad@2x.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFullVirtualPath(tt.carPath, tt.asset)
			assert.Equal(t, tt.expected, result)
		})
	}
}
