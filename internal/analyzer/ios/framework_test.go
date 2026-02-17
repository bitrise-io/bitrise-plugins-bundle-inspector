package ios

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverFrameworks(t *testing.T) {
	wikipediaApp := "../../../test-artifacts/ios/Wikipedia.app"

	if _, err := os.Stat(wikipediaApp); os.IsNotExist(err) {
		t.Skip("Wikipedia test artifact not found")
	}

	frameworks, err := DiscoverFrameworks(wikipediaApp)
	require.NoError(t, err)
	assert.NotEmpty(t, frameworks, "Should find at least one framework")

	// Check that WMF.framework is found
	foundWMF := false
	for _, fw := range frameworks {
		if fw.Name == "WMF.framework" {
			foundWMF = true
			assert.Equal(t, "Frameworks/WMF.framework", fw.Path)
			assert.NotEmpty(t, fw.Version, "Should have version")
			assert.True(t, fw.Size > 0, "Should have size")
			assert.NotNil(t, fw.BinaryInfo, "Should have binary info")
			assert.Equal(t, "arm64", fw.BinaryInfo.Architecture)
			assert.Equal(t, "dylib", fw.BinaryInfo.Type)
			assert.NotEmpty(t, fw.Dependencies, "Should have dependencies")
			break
		}
	}
	assert.True(t, foundWMF, "Should find WMF.framework")
}

func TestParseFrameworkInfo(t *testing.T) {
	frameworkPath := "../../../test-artifacts/ios/Wikipedia.app/Frameworks/WMF.framework"
	appPath := "../../../test-artifacts/ios/Wikipedia.app"

	if _, err := os.Stat(frameworkPath); os.IsNotExist(err) {
		t.Skip("WMF framework test artifact not found")
	}

	info, err := ParseFrameworkInfo(frameworkPath, appPath)
	require.NoError(t, err)
	assert.Equal(t, "WMF.framework", info.Name)
	assert.NotEmpty(t, info.Version, "Should have version")
	assert.True(t, info.Size > 0, "Should have size")
	assert.NotNil(t, info.BinaryInfo, "Should have binary info")
}

func TestGetFrameworkVersion(t *testing.T) {
	infoPlistPath := "../../../test-artifacts/ios/Wikipedia.app/Frameworks/WMF.framework/Info.plist"

	if _, err := os.Stat(infoPlistPath); os.IsNotExist(err) {
		t.Skip("Info.plist test artifact not found")
	}

	version, err := GetFrameworkVersion(infoPlistPath)
	require.NoError(t, err)
	assert.NotEmpty(t, version, "Should have version")
	t.Logf("WMF framework version: %s", version)
}

func TestExtractIconNames(t *testing.T) {
	tests := []struct {
		name     string
		plist    map[string]interface{}
		expected []string
	}{
		{
			name: "standard CFBundleIcons with CFBundleIconFiles",
			plist: map[string]interface{}{
				"CFBundleIcons": map[string]interface{}{
					"CFBundlePrimaryIcon": map[string]interface{}{
						"CFBundleIconFiles": []interface{}{"AppIcon60x60"},
					},
				},
			},
			expected: []string{"AppIcon60x60"},
		},
		{
			name: "CFBundleIconName alongside CFBundleIconFiles",
			plist: map[string]interface{}{
				"CFBundleIcons": map[string]interface{}{
					"CFBundlePrimaryIcon": map[string]interface{}{
						"CFBundleIconFiles": []interface{}{"AppIcon60x60", "AppIcon76x76"},
						"CFBundleIconName":  "AppIcon",
					},
				},
			},
			expected: []string{"AppIcon60x60", "AppIcon76x76", "AppIcon"},
		},
		{
			name: "iPad variant CFBundleIcons~ipad",
			plist: map[string]interface{}{
				"CFBundleIcons~ipad": map[string]interface{}{
					"CFBundlePrimaryIcon": map[string]interface{}{
						"CFBundleIconFiles": []interface{}{"AppIcon76x76"},
						"CFBundleIconName":  "AppIcon",
					},
				},
			},
			expected: []string{"AppIcon76x76", "AppIcon"},
		},
		{
			name: "both iPhone and iPad variants with deduplication",
			plist: map[string]interface{}{
				"CFBundleIcons": map[string]interface{}{
					"CFBundlePrimaryIcon": map[string]interface{}{
						"CFBundleIconFiles": []interface{}{"AppIcon60x60"},
						"CFBundleIconName":  "AppIcon",
					},
				},
				"CFBundleIcons~ipad": map[string]interface{}{
					"CFBundlePrimaryIcon": map[string]interface{}{
						"CFBundleIconFiles": []interface{}{"AppIcon76x76"},
						"CFBundleIconName":  "AppIcon",
					},
				},
			},
			expected: []string{"AppIcon60x60", "AppIcon", "AppIcon76x76"},
		},
		{
			name: "legacy top-level CFBundleIconFiles with .png extension stripped",
			plist: map[string]interface{}{
				"CFBundleIconFiles": []interface{}{"Icon.png", "Icon@2x.png"},
			},
			expected: []string{"Icon", "Icon@2x"},
		},
		{
			name: "legacy singular CFBundleIconFile with .png extension stripped",
			plist: map[string]interface{}{
				"CFBundleIconFile": "Icon.png",
			},
			expected: []string{"Icon"},
		},
		{
			name: "custom icon names (Facebook-style)",
			plist: map[string]interface{}{
				"CFBundleIcons": map[string]interface{}{
					"CFBundlePrimaryIcon": map[string]interface{}{
						"CFBundleIconFiles": []interface{}{"Icon-Production"},
						"CFBundleIconName":  "Icon-Production",
					},
				},
			},
			expected: []string{"Icon-Production"},
		},
		{
			name:     "empty plist - no icon keys",
			plist:    map[string]interface{}{},
			expected: nil,
		},
		{
			name: "CFBundleIcons without CFBundlePrimaryIcon",
			plist: map[string]interface{}{
				"CFBundleIcons": map[string]interface{}{},
			},
			expected: nil,
		},
		{
			name: "empty string values are skipped",
			plist: map[string]interface{}{
				"CFBundleIconFile": "",
				"CFBundleIcons": map[string]interface{}{
					"CFBundlePrimaryIcon": map[string]interface{}{
						"CFBundleIconName": "",
					},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIconNames(tt.plist)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAppInfoPlist_IconNames(t *testing.T) {
	infoPlistPath := filepath.Join("..", "..", "..", "test-artifacts", "ios", "Wikipedia.app", "Info.plist")

	if _, err := os.Stat(infoPlistPath); os.IsNotExist(err) {
		t.Skip("Wikipedia Info.plist test artifact not found")
	}

	metadata, err := ParseAppInfoPlist(infoPlistPath)
	require.NoError(t, err)
	require.NotNil(t, metadata)

	// Wikipedia.app should have icon names declared in Info.plist
	t.Logf("Icon names: %v", metadata.IconNames)
	// Even if empty, this shouldn't error - just means icons are in Assets.car only
}
