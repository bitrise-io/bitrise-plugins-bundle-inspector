//go:build !cgo || !darwin
// +build !cgo !darwin

package compression

import (
	"bytes"
	"fmt"
)

// LZFSE compression method code used in ZIP files
const CompressionMethodLZFSE = 99

// LZFSE magic bytes
var lzfseMagic = [][]byte{
	[]byte("bvx-"), // LZFSE compressed block
	[]byte("bvx$"), // LZFSE end of stream
	[]byte("bvxn"), // LZFSE uncompressed block
	[]byte("bvx2"), // LZVN compressed block
}

// DecompressLZFSE decompresses LZFSE-compressed data.
// This is a stub implementation for non-CGo builds that returns an error.
func DecompressLZFSE(compressed []byte) ([]byte, error) {
	return nil, fmt.Errorf("LZFSE decompression not supported in this build (requires CGo and macOS)")
}

// IsLZFSECompressed checks if data starts with LZFSE magic bytes.
func IsLZFSECompressed(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	header := data[:4]
	for _, magic := range lzfseMagic {
		if bytes.Equal(header, magic) {
			return true
		}
	}

	return false
}
