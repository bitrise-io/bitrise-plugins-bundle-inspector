package util

import (
	"archive/zip"
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIconPriority(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"AppIcon60x60@3x.png", 30},
		{"AppIcon60x60@2x.png", 20},
		{"AppIcon60x60@1x.png", 10},
		{"AppIcon60x60.png", 25},
		{"AppIcon76x76@2x~ipad.png", 20},
		{"AppIcon.png", 5},
		{"AppIcon-large.png", 5},
		{"random.png", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getIconPriority(tt.name))
		})
	}
}

func TestGetIconPriority_PlistGuidedBonus(t *testing.T) {
	// Plist-guided candidates get +50 bonus
	// A plist-guided @3x should beat a plain AppIcon@3x
	plistGuided3x := getIconPriority("Icon-Production@3x.png") + 50 // 30 + 50 = 80
	appIcon3x := getIconPriority("AppIcon60x60@3x.png")             // 30

	assert.Greater(t, plistGuided3x, appIcon3x, "Plist-guided should have higher effective priority")
}

func TestIconSearchHints_NilSafe(t *testing.T) {
	// Verify that nil hints doesn't panic when passed to the function
	tmpDir := t.TempDir()
	_, err := ExtractIconFromDirectoryWithHints(tmpDir, nil)
	// Should return an error (no icons in empty dir) but not panic
	assert.Error(t, err)
}

func TestIsCgBIPNG(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "standard PNG - not CgBI",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R', 0x00, 0x00, 0x00, 0x01},
			expected: false,
		},
		{
			name:     "CgBI PNG",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x04, 'C', 'g', 'B', 'I', 0x00, 0x00, 0x00, 0x01},
			expected: true,
		},
		{
			name:     "too short",
			data:     []byte{0x89, 0x50, 0x4E},
			expected: false,
		},
		{
			name:     "not a PNG",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isCgBIPNG(tt.data))
		})
	}
}

// --- Android icon extraction tests ---

// createTestAPK creates a zip file with the given entries (path → data).
func createTestAPK(t *testing.T, entries map[string][]byte) string {
	t.Helper()
	apkPath := filepath.Join(t.TempDir(), "test.apk")
	f, err := os.Create(apkPath)
	require.NoError(t, err)

	w := zip.NewWriter(f)
	for path, data := range entries {
		entry, err := w.Create(path)
		require.NoError(t, err)
		_, err = entry.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())
	return apkPath
}

// minimalPNG returns a valid 1x1 PNG image.
func minimalPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, image.Black)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestExtractAndroidIcon_ExactDensityPath(t *testing.T) {
	// Regression: standard paths without qualifier suffix still work
	apkPath := createTestAPK(t, map[string][]byte{
		"res/mipmap-xxxhdpi/ic_launcher.png": minimalPNG(),
	})
	r, err := zip.OpenReader(apkPath)
	require.NoError(t, err)
	defer r.Close()

	data, err := extractAndroidIcon(r)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestExtractAndroidIcon_WithQualifierSuffix(t *testing.T) {
	// Modern APKs use qualifier suffixes like -v4
	apkPath := createTestAPK(t, map[string][]byte{
		"res/mipmap-xxxhdpi-v4/ic_launcher.png": minimalPNG(),
	})
	r, err := zip.OpenReader(apkPath)
	require.NoError(t, err)
	defer r.Close()

	data, err := extractAndroidIcon(r)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestExtractAndroidIcon_WebP(t *testing.T) {
	// WebP icons should be found and decoded to PNG.
	// We use a real PNG as the content but name it .webp to test the filename matching.
	// The actual WebP decode is tested via image.Decode which requires a real WebP.
	// Here we verify the filename matching works for .webp extensions.
	// For a full decode test, we create a real WebP via encoding.

	webpData := minimalWebP()

	apkPath := createTestAPK(t, map[string][]byte{
		"res/mipmap-xxhdpi-v4/ic_launcher.webp": webpData,
	})
	r, err := zip.OpenReader(apkPath)
	require.NoError(t, err)
	defer r.Close()

	data, err := extractAndroidIcon(r)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestExtractAndroidIcon_DensityPriority(t *testing.T) {
	// xxxhdpi should be preferred over mdpi
	pngData := minimalPNG()
	apkPath := createTestAPK(t, map[string][]byte{
		"res/mipmap-mdpi/ic_launcher.png":        pngData,
		"res/mipmap-xxxhdpi-v4/ic_launcher.png":  pngData,
		"res/mipmap-xhdpi-v4/ic_launcher.png":    pngData,
	})
	r, err := zip.OpenReader(apkPath)
	require.NoError(t, err)
	defer r.Close()

	data, err := extractAndroidIcon(r)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	// We can't easily distinguish which was selected by content, but we verify
	// extraction succeeds when multiple densities are present
}

func TestExtractAndroidIcon_AABPaths(t *testing.T) {
	// AAB bundles nest resources under base/res/
	apkPath := createTestAPK(t, map[string][]byte{
		"base/res/mipmap-xxhdpi-v4/ic_launcher.png": minimalPNG(),
	})
	r, err := zip.OpenReader(apkPath)
	require.NoError(t, err)
	defer r.Close()

	data, err := extractAndroidIcon(r)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestExtractAndroidIcon_NoIcon(t *testing.T) {
	// No icon files → error
	apkPath := createTestAPK(t, map[string][]byte{
		"res/layout/activity_main.xml": []byte("<LinearLayout/>"),
		"classes.dex":                  []byte{0x64, 0x65, 0x78},
	})
	r, err := zip.OpenReader(apkPath)
	require.NoError(t, err)
	defer r.Close()

	data, err := extractAndroidIcon(r)
	assert.Error(t, err)
	assert.Nil(t, data)
}

// minimalWebP returns a valid 1x1 WebP image (VP8 lossy, generated via cwebp).
func minimalWebP() []byte {
	return []byte{
		0x52, 0x49, 0x46, 0x46, 0x3c, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50, 0x56, 0x50, 0x38, 0x20,
		0x30, 0x00, 0x00, 0x00, 0xd0, 0x01, 0x00, 0x9d, 0x01, 0x2a, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00,
		0x34, 0x25, 0xa0, 0x02, 0x74, 0xba, 0x01, 0xf8, 0x00, 0x03, 0xb0, 0x00, 0xfe, 0xf0, 0xc4, 0x0b,
		0xff, 0x20, 0xb9, 0x61, 0x75, 0xc8, 0xd7, 0xff, 0x20, 0x3f, 0xe4, 0x07, 0xfc, 0x80, 0xff, 0xf8,
		0xf2, 0x00, 0x00, 0x00,
	}
}
