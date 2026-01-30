// Package util provides utility functions for the bundle inspector.
package util

// BlockSize is the standard block size for iOS app bundles (4 KB)
// Files in IPAs and on iOS devices are aligned to 4 KB boundaries
// iOS uses APFS with 4 KB block size, so even small files occupy a full block
const BlockSize = 4096

// CalculateDiskUsage returns the actual disk space used by a file.
// iOS uses APFS with 4 KB block size, so even small files occupy a full block.
// This rounds up the file size to the nearest block boundary.
func CalculateDiskUsage(fileSize int64) int64 {
	if fileSize == 0 {
		return 0
	}
	// Round up to nearest block size
	blocks := (fileSize + BlockSize - 1) / BlockSize
	return blocks * BlockSize
}
