package compression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLZFSECompressed(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "LZFSE compressed block",
			data:     []byte("bvx-data here"),
			expected: true,
		},
		{
			name:     "LZFSE end of stream",
			data:     []byte("bvx$data here"),
			expected: true,
		},
		{
			name:     "LZFSE uncompressed block",
			data:     []byte("bvxndata here"),
			expected: true,
		},
		{
			name:     "LZVN compressed block",
			data:     []byte("bvx2data here"),
			expected: true,
		},
		{
			name:     "ZIP magic",
			data:     []byte("PK\x03\x04"),
			expected: false,
		},
		{
			name:     "Random data",
			data:     []byte("random data"),
			expected: false,
		},
		{
			name:     "Too short",
			data:     []byte("bv"),
			expected: false,
		},
		{
			name:     "Empty",
			data:     []byte{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLZFSECompressed(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDecompressLZFSE(t *testing.T) {
	// Test with empty data
	_, err := DecompressLZFSE([]byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty compressed data")

	// Test with invalid LZFSE data
	_, err = DecompressLZFSE([]byte("not lzfse data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LZFSE decompression failed")

	// Note: Testing with real LZFSE compressed data would require
	// having a sample compressed file, which we don't have in test artifacts.
	// The actual decompression is tested when processing real IPA files.
}

func TestCompressionMethodConstant(t *testing.T) {
	// Verify the LZFSE compression method constant
	assert.Equal(t, 99, CompressionMethodLZFSE)
}
