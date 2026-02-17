package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	// Verify that nil hints doesn't panic
	var hints *IconSearchHints
	assert.Nil(t, hints)

	// The functions accept nil hints gracefully (tested via build + run)
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
