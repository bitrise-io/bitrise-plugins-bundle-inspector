package detector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathAnalyzer_GetBundleBoundaries(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name           string
		path           string
		wantInBundle   bool
		wantBundleName string
		wantBundleType string
	}{
		{
			name:           "Framework bundle",
			path:           "Payload/App.app/Frameworks/GoogleMaps.framework/GoogleMaps",
			wantInBundle:   true,
			wantBundleName: "GoogleMaps.framework",
			wantBundleType: "framework",
		},
		{
			name:           "Framework Info.plist",
			path:           "Payload/App.app/Frameworks/AFNetworking.framework/Info.plist",
			wantInBundle:   true,
			wantBundleName: "AFNetworking.framework",
			wantBundleType: "framework",
		},
		{
			name:           "Extension bundle",
			path:           "Payload/App.app/PlugIns/ShareExtension.appex/ShareExtension",
			wantInBundle:   true,
			wantBundleName: "ShareExtension.appex",
			wantBundleType: "appex",
		},
		{
			name:           "App bundle",
			path:           "Payload/MyApp.app/MyApp",
			wantInBundle:   true,
			wantBundleName: "MyApp.app",
			wantBundleType: "app",
		},
		{
			name:           "Asset catalog",
			path:           "Payload/App.app/Assets.xcassets/AppIcon.appiconset/Contents.json",
			wantInBundle:   true,
			wantBundleName: "Assets.xcassets",
			wantBundleType: "xcassets",
		},
		{
			name:           "Localization bundle",
			path:           "Payload/App.app/en.lproj/Localizable.strings",
			wantInBundle:   true,
			wantBundleName: "en.lproj",
			wantBundleType: "lproj",
		},
		{
			name:           "Framework localization",
			path:           "Payload/App.app/Frameworks/SDK.framework/en.lproj/strings.strings",
			wantInBundle:   true,
			wantBundleName: "en.lproj",
			wantBundleType: "lproj",
		},
		{
			name:           "File in app bundle",
			path:           "Payload/App.app/image.png",
			wantInBundle:   true,
			wantBundleName: "App.app",
			wantBundleType: "app",
		},
		{
			name:         "Root level file",
			path:         "config.json",
			wantInBundle: false,
		},
		{
			name:           "Nested frameworks (detects innermost)",
			path:           "Payload/App.app/Frameworks/Outer.framework/Frameworks/Inner.framework/file.txt",
			wantInBundle:   true,
			wantBundleName: "Inner.framework",
			wantBundleType: "framework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := analyzer.GetBundleBoundaries(tt.path)
			assert.Equal(t, tt.wantInBundle, info.InBundle, "InBundle mismatch")
			if tt.wantInBundle {
				assert.Equal(t, tt.wantBundleName, info.Name, "Bundle name mismatch")
				assert.Equal(t, tt.wantBundleType, info.Type, "Bundle type mismatch")
			}
		})
	}
}

func TestPathAnalyzer_IsFrameworkPath(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Framework path",
			path: "Payload/App.app/Frameworks/GoogleMaps.framework/GoogleMaps",
			want: true,
		},
		{
			name: "Framework resource",
			path: "Payload/App.app/Frameworks/SDK.framework/Resources/icon.png",
			want: true,
		},
		{
			name: "Not a framework",
			path: "Payload/App.app/MyApp",
			want: false,
		},
		{
			name: "Extension (not framework)",
			path: "Payload/App.app/PlugIns/Widget.appex/Widget",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.IsFrameworkPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_IsExtensionPath(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Extension path",
			path: "Payload/App.app/PlugIns/ShareExtension.appex/ShareExtension",
			want: true,
		},
		{
			name: "Extension resource",
			path: "Payload/App.app/PlugIns/Widget.appex/Assets.car",
			want: true,
		},
		{
			name: "Not an extension",
			path: "Payload/App.app/MyApp",
			want: false,
		},
		{
			name: "Framework (not extension)",
			path: "Payload/App.app/Frameworks/SDK.framework/SDK",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.IsExtensionPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_IsAssetCatalogPath(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "xcassets path",
			path: "Payload/App.app/Assets.xcassets/AppIcon.appiconset/Contents.json",
			want: true,
		},
		{
			name: "Assets.car file",
			path: "Payload/App.app/Assets.car",
			want: true,
		},
		{
			name: "Assets.car extraction",
			path: "Payload/App.app/Assets.car/AppIcon60x60@2x.png",
			want: true,
		},
		{
			name: "Not asset catalog",
			path: "Payload/App.app/image.png",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.IsAssetCatalogPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_IsLocalizationPath(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Localization bundle",
			path: "Payload/App.app/en.lproj/Localizable.strings",
			want: true,
		},
		{
			name: "Framework localization",
			path: "Payload/App.app/Frameworks/SDK.framework/de.lproj/strings.strings",
			want: true,
		},
		{
			name: "Not localization",
			path: "Payload/App.app/file.strings",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.IsLocalizationPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_ExtractFrameworkName(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Framework path",
			path: "Payload/App.app/Frameworks/GoogleMaps.framework/GoogleMaps",
			want: "GoogleMaps",
		},
		{
			name: "Framework resource",
			path: "Payload/App.app/Frameworks/AFNetworking.framework/Info.plist",
			want: "AFNetworking",
		},
		{
			name: "No framework",
			path: "Payload/App.app/MyApp",
			want: "",
		},
		{
			name: "Framework with version",
			path: "Frameworks/Firebase.framework/Firebase",
			want: "Firebase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.ExtractFrameworkName(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_ExtractExtensionName(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Extension path",
			path: "Payload/App.app/PlugIns/ShareExtension.appex/ShareExtension",
			want: "ShareExtension",
		},
		{
			name: "Extension resource",
			path: "Payload/App.app/PlugIns/Widget.appex/Info.plist",
			want: "Widget",
		},
		{
			name: "No extension",
			path: "Payload/App.app/MyApp",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.ExtractExtensionName(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_GetDistinctBundles(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name  string
		paths []string
		want  int // Number of distinct bundles
	}{
		{
			name: "Two different frameworks",
			paths: []string{
				"Payload/App.app/Frameworks/A.framework/A",
				"Payload/App.app/Frameworks/B.framework/B",
			},
			want: 2,
		},
		{
			name: "Same framework multiple files",
			paths: []string{
				"Payload/App.app/Frameworks/A.framework/A",
				"Payload/App.app/Frameworks/A.framework/Info.plist",
				"Payload/App.app/Frameworks/A.framework/Resources/icon.png",
			},
			want: 1,
		},
		{
			name: "Mixed bundle types",
			paths: []string{
				"Payload/App.app/Frameworks/SDK.framework/SDK",
				"Payload/App.app/PlugIns/Share.appex/Share",
				"Payload/App.app/Assets.xcassets/AppIcon.appiconset/Contents.json",
			},
			want: 3,
		},
		{
			name: "No bundles",
			paths: []string{
				"Payload/image.png",
				"Payload/config.json",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundles := analyzer.GetDistinctBundles(tt.paths)
			assert.Equal(t, tt.want, len(bundles))
		})
	}
}

func TestPathAnalyzer_AreInDifferentBundles(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name  string
		path1 string
		path2 string
		want  bool
	}{
		{
			name:  "Different frameworks",
			path1: "Payload/App.app/Frameworks/A.framework/Info.plist",
			path2: "Payload/App.app/Frameworks/B.framework/Info.plist",
			want:  true,
		},
		{
			name:  "Same framework",
			path1: "Payload/App.app/Frameworks/A.framework/A",
			path2: "Payload/App.app/Frameworks/A.framework/Info.plist",
			want:  false,
		},
		{
			name:  "Different bundle types",
			path1: "Payload/App.app/Frameworks/A.framework/file.txt",
			path2: "Payload/App.app/PlugIns/Share.appex/file.txt",
			want:  false, // Different types don't count
		},
		{
			name:  "One not in bundle",
			path1: "Payload/App.app/Frameworks/A.framework/file.txt",
			path2: "Payload/App.app/file.txt",
			want:  false,
		},
		{
			name:  "Neither in bundle",
			path1: "Payload/App.app/file1.txt",
			path2: "Payload/App.app/file2.txt",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.AreInDifferentBundles(tt.path1, tt.path2)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_GetFileName(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Simple filename",
			path: "Payload/App.app/Info.plist",
			want: "Info.plist",
		},
		{
			name: "Deep path",
			path: "Payload/App.app/Frameworks/SDK.framework/Resources/icon.png",
			want: "icon.png",
		},
		{
			name: "Root file",
			path: "file.txt",
			want: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.GetFileName(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_GetFileExtension(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "plist file",
			path: "Payload/App.app/Info.plist",
			want: "plist",
		},
		{
			name: "png file",
			path: "icon.png",
			want: "png",
		},
		{
			name: "No extension",
			path: "Payload/App.app/MyApp",
			want: "",
		},
		{
			name: "Multiple dots",
			path: "file.backup.json",
			want: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.GetFileExtension(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPathAnalyzer_ExtractLocaleDirectory(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name       string
		path       string
		wantLocale string
		wantFound  bool
	}{
		{
			name:       "Spanish region variant",
			path:       "Resources/es_419/messages.offline_catalog",
			wantLocale: "es_419",
			wantFound:  true,
		},
		{
			name:       "Hebrew legacy alias",
			path:       "Resources/he_ALL/messages.offline_catalog",
			wantLocale: "he_ALL",
			wantFound:  true,
		},
		{
			name:       "Chinese variant with dash",
			path:       "Resources/zh-CN_ALL/data.bin",
			wantLocale: "zh-CN_ALL",
			wantFound:  true,
		},
		{
			name:       "Standard locale",
			path:       "Resources/en_US/strings.bin",
			wantLocale: "en_US",
			wantFound:  true,
		},
		{
			name:       "No locale directory",
			path:       "Payload/App.app/Resources/data.bin",
			wantLocale: "",
			wantFound:  false,
		},
		{
			name:       "Single letter directory (not locale)",
			path:       "Resources/a/data.bin",
			wantLocale: "",
			wantFound:  false,
		},
		{
			name:       "lproj is not a locale directory",
			path:       "Resources/en.lproj/Localizable.strings",
			wantLocale: "",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locale, found := analyzer.ExtractLocaleDirectory(tt.path)
			assert.Equal(t, tt.wantFound, found, "Found mismatch")
			assert.Equal(t, tt.wantLocale, locale, "Locale mismatch")
		})
	}
}

func TestPathAnalyzer_ReplaceLocaleDirectory(t *testing.T) {
	analyzer := NewPathAnalyzer("")

	tests := []struct {
		name           string
		path           string
		wantNormalized string
		wantFound      bool
	}{
		{
			name:           "Replace Spanish variant",
			path:           "Resources/es_419/messages.offline_catalog",
			wantNormalized: "Resources/<LOCALE>/messages.offline_catalog",
			wantFound:      true,
		},
		{
			name:           "Replace Chinese variant",
			path:           "Resources/zh-CN_ALL/data.bin",
			wantNormalized: "Resources/<LOCALE>/data.bin",
			wantFound:      true,
		},
		{
			name:           "No locale to replace",
			path:           "Payload/App.app/data.bin",
			wantNormalized: "Payload/App.app/data.bin",
			wantFound:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, found := analyzer.ReplaceLocaleDirectory(tt.path)
			assert.Equal(t, tt.wantFound, found, "Found mismatch")
			assert.Equal(t, tt.wantNormalized, normalized, "Normalized mismatch")
		})
	}
}
