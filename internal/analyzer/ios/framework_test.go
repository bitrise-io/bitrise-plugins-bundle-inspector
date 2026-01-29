package ios

import (
	"os"
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
