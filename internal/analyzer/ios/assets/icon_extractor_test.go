package assets

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractIconFromCar_InvalidPath(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Assets.car icon extraction requires macOS")
	}

	_, err := ExtractIconFromCar("/nonexistent/Assets.car", []string{"AppIcon"}, nil)
	assert.Error(t, err)
}

func TestExtractIconFromCar_EmptyCandidates(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Assets.car icon extraction requires macOS")
	}

	// Create a temp file that's not a real .car
	tmpFile, err := os.CreateTemp("", "fake-*.car")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("not a real car file"))
	tmpFile.Close()

	_, err = ExtractIconFromCar(tmpFile.Name(), nil, nil)
	assert.Error(t, err)
	// With nil iconNames and nil catalogAssets, only fallback names are tried
	assert.Contains(t, err.Error(), "swift icon extraction failed")
}

func TestExtractIconFromCar_WithCatalogAssets(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Assets.car icon extraction requires macOS")
	}

	// Test that catalog assets with type "icon" are used as candidates
	catalogAssets := []AssetInfo{
		{Name: "CustomIcon", Type: "icon"},
		{Name: "SomeImage", Type: "image"},
		{Name: "AnotherIcon", Type: "icon"},
	}

	// With a fake .car file, extraction will fail but should try the right names
	tmpFile, err := os.CreateTemp("", "fake-*.car")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("not a real car file"))
	tmpFile.Close()

	_, err = ExtractIconFromCar(tmpFile.Name(), []string{"PlistIcon"}, catalogAssets)
	assert.Error(t, err)
	// All 6 candidates are passed to a single Swift invocation
	assert.Contains(t, err.Error(), "swift icon extraction failed")
}

func TestExtractIconFromCar_Wikipedia(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Assets.car icon extraction requires macOS")
	}

	carPath := filepath.Join("..", "..", "..", "..", "test-artifacts", "ios", "Wikipedia.app", "Assets.car")
	if _, err := os.Stat(carPath); os.IsNotExist(err) {
		t.Skip("Wikipedia Assets.car test artifact not found")
	}

	data, err := ExtractIconFromCar(carPath, []string{"AppIcon"}, nil)
	require.NoError(t, err)
	assert.True(t, len(data) > 100, "Icon data should be non-trivial in size")

	// Verify PNG magic bytes
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	assert.Equal(t, pngMagic, data[:8], "Output should be a valid PNG")
}
