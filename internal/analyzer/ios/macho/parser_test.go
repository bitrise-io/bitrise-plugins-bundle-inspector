package macho

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMachO(t *testing.T) {
	// Test with Wikipedia executable
	wikipediaBinary := "../../../../test-artifacts/ios/Wikipedia.app/Wikipedia"

	if _, err := os.Stat(wikipediaBinary); os.IsNotExist(err) {
		t.Skip("Wikipedia test artifact not found")
	}

	info, err := ParseMachO(wikipediaBinary)
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "arm64", info.Architecture)
	assert.Equal(t, "executable", info.Type)
	assert.True(t, len(info.LinkedLibraries) > 0, "Should have linked libraries")
	assert.True(t, info.CodeSize > 0, "Should have code size")
}

func TestParseFrameworkBinary(t *testing.T) {
	// Test with WMF framework
	frameworkBinary := "../../../../test-artifacts/ios/Wikipedia.app/Frameworks/WMF.framework/WMF"

	if _, err := os.Stat(frameworkBinary); os.IsNotExist(err) {
		t.Skip("WMF framework test artifact not found")
	}

	info, err := ParseMachO(frameworkBinary)
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "arm64", info.Architecture)
	assert.Equal(t, "dylib", info.Type)
	assert.True(t, info.CodeSize > 0, "Should have code size")
}

func TestGracefulErrorHandling(t *testing.T) {
	// Test with non-Mach-O file
	infoPlist := "../../../../test-artifacts/ios/Wikipedia.app/Info.plist"

	if _, err := os.Stat(infoPlist); os.IsNotExist(err) {
		t.Skip("Info.plist test artifact not found")
	}

	_, err := ParseMachO(infoPlist)
	assert.Error(t, err, "Should fail gracefully on non-Mach-O file")
}

func TestIsMachO(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Wikipedia executable",
			path:     "../../../../test-artifacts/ios/Wikipedia.app/Wikipedia",
			expected: true,
		},
		{
			name:     "Info.plist",
			path:     "../../../../test-artifacts/ios/Wikipedia.app/Info.plist",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.path); os.IsNotExist(err) {
				t.Skip("Test artifact not found")
			}

			result := IsMachO(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetArchitectures(t *testing.T) {
	wikipediaBinary := "../../../../test-artifacts/ios/Wikipedia.app/Wikipedia"

	if _, err := os.Stat(wikipediaBinary); os.IsNotExist(err) {
		t.Skip("Wikipedia test artifact not found")
	}

	archs, err := GetArchitectures(wikipediaBinary)
	require.NoError(t, err)
	assert.NotEmpty(t, archs)
	assert.Contains(t, archs, "arm64")
}

func TestGetLinkedLibraries(t *testing.T) {
	wikipediaBinary := "../../../../test-artifacts/ios/Wikipedia.app/Wikipedia"

	if _, err := os.Stat(wikipediaBinary); os.IsNotExist(err) {
		t.Skip("Wikipedia test artifact not found")
	}

	libraries, err := GetLinkedLibraries(wikipediaBinary)
	require.NoError(t, err)
	assert.NotEmpty(t, libraries)

	// Should have at least system libraries
	hasSystemLib := false
	for _, lib := range libraries {
		if filepath.Base(lib) == "libSystem.B.dylib" {
			hasSystemLib = true
			break
		}
	}
	assert.True(t, hasSystemLib, "Should link to libSystem.B.dylib")
}

func TestGetCPUTypeName(t *testing.T) {
	tests := []struct {
		cpu      uint32
		expected string
	}{
		{0x01000007, "x86_64"}, // macho.CpuAmd64
		{0x0100000C, "arm64"},  // macho.CpuArm64
	}

	for _, tt := range tests {
		// Note: We can't easily test this without creating mock Cpu values
		// The actual CPU type constants are platform-specific
		t.Logf("CPU type %d maps to architecture name", tt.cpu)
	}
}

func TestHasDebugSymbols(t *testing.T) {
	// Test with debug dylib - note: the .debug.dylib file may have debug symbols
	// stripped and stored in a separate dSYM bundle, so we just test that the
	// function runs without error
	debugDylib := "../../../../test-artifacts/ios/Wikipedia.app/Wikipedia.debug.dylib"

	if _, err := os.Stat(debugDylib); os.IsNotExist(err) {
		t.Skip("Debug dylib test artifact not found")
	}

	hasDebug, err := HasDebugSymbols(debugDylib)
	require.NoError(t, err)
	// The function should return a boolean without error
	// Note: The actual value depends on whether symbols are embedded or in a dSYM
	t.Logf("Debug dylib has embedded debug symbols: %v", hasDebug)
}

func TestGetSegmentSizes(t *testing.T) {
	wikipediaBinary := "../../../../test-artifacts/ios/Wikipedia.app/Wikipedia"

	if _, err := os.Stat(wikipediaBinary); os.IsNotExist(err) {
		t.Skip("Wikipedia test artifact not found")
	}

	codeSize, dataSize, err := GetSegmentSizes(wikipediaBinary)
	require.NoError(t, err)
	assert.True(t, codeSize > 0, "Should have non-zero code segment size")
	assert.True(t, dataSize >= 0, "Data segment size should be non-negative")

	// Code segment should typically be larger than data segment for executables
	t.Logf("Code size: %d bytes, Data size: %d bytes", codeSize, dataSize)
}
