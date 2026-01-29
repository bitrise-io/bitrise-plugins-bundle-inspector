package compression

import (
	"bytes"
	"fmt"

	"github.com/blacktop/lzfse-cgo"
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
func DecompressLZFSE(compressed []byte) ([]byte, error) {
	if len(compressed) == 0 {
		return nil, fmt.Errorf("empty compressed data")
	}

	// Use blacktop/lzfse-cgo for decompression
	decompressed := lzfse.DecodeBuffer(compressed)

	// Check if decompression succeeded
	if len(decompressed) == 0 {
		return nil, fmt.Errorf("LZFSE decompression failed: output is empty")
	}

	return decompressed, nil
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
