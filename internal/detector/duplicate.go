// Package detector provides file analysis and detection capabilities.
package detector

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

const (
	// blockSize is the standard block size for iOS app bundles (4 KB)
	// Files in IPAs and on iOS devices are aligned to 4 KB boundaries
	blockSize = 4096
)

// DuplicateDetector detects duplicate files using SHA-256 hashing.
type DuplicateDetector struct {
	// Map of size -> list of file paths
	sizeGroups map[int64][]string
	// Map of hash -> list of file paths
	hashGroups map[string][]string
}

// NewDuplicateDetector creates a new duplicate detector.
func NewDuplicateDetector() *DuplicateDetector {
	return &DuplicateDetector{
		sizeGroups: make(map[int64][]string),
		hashGroups: make(map[string][]string),
	}
}

// DetectDuplicates finds duplicate files in the extracted artifact directory.
func (d *DuplicateDetector) DetectDuplicates(rootPath string) ([]types.DuplicateSet, error) {
	// Phase 1: Group files by size (cheap operation)
	if err := d.groupBySize(rootPath); err != nil {
		return nil, err
	}

	// Phase 2: For size collisions, compute hashes
	d.hashGroups = make(map[string][]string)

	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make(chan error, 1)

	// Process size groups that have multiple files
	for size, files := range d.sizeGroups {
		if len(files) < 2 {
			continue
		}

		// Hash all files of this size in parallel
		for _, file := range files {
			wg.Add(1)
			go func(path string, size int64) {
				defer wg.Done()

				hash, err := util.ComputeSHA256(path)
				if err != nil {
					select {
					case errors <- err:
					default:
					}
					return
				}

				mu.Lock()
				d.hashGroups[hash] = append(d.hashGroups[hash], path)
				mu.Unlock()
			}(file, size)
		}
	}

	wg.Wait()
	close(errors)

	// Check for errors
	if err := <-errors; err != nil {
		return nil, err
	}

	// Phase 3: Build duplicate sets from hash groups
	var duplicates []types.DuplicateSet
	for hash, files := range d.hashGroups {
		if len(files) < 2 {
			continue
		}

		// Get file size (all files in group have same size)
		info, err := os.Stat(files[0])
		if err != nil {
			continue
		}

		// Use block-aligned size for accurate space calculations
		// iOS apps store files in 4 KB blocks, so even small files occupy a full block
		actualSize := info.Size()
		alignedSize := blockAlignedSize(actualSize)

		dup := types.DuplicateSet{
			Hash:       hash,
			Size:       alignedSize,          // Report block-aligned size
			Count:      len(files),
			Files:      files,
			WastedSize: (int64(len(files)) - 1) * alignedSize, // Calculate waste based on aligned size
		}
		duplicates = append(duplicates, dup)
	}

	return duplicates, nil
}

// shouldSkipFile returns true if the file should be excluded from duplicate detection.
// Some files are legitimately duplicated by design and shouldn't be reported.
func shouldSkipFile(path string) bool {
	filename := filepath.Base(path)

	// Exclude files that are required to be separate per framework/component
	excludePatterns := []string{
		"PrivacyInfo.xcprivacy", // Apple privacy manifests - required per framework by Apple
	}

	for _, pattern := range excludePatterns {
		if filename == pattern {
			return true
		}
	}

	return false
}

// groupBySize groups files by their size.
func (d *DuplicateDetector) groupBySize(rootPath string) error {
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			return nil
		}

		// Skip files that are legitimately duplicated by design
		if shouldSkipFile(path) {
			return nil
		}

		// Group by size
		size := info.Size()
		d.sizeGroups[size] = append(d.sizeGroups[size], path)

		return nil
	})
}

// GetTotalWastedSpace calculates total wasted space from duplicates.
func GetTotalWastedSpace(duplicates []types.DuplicateSet) int64 {
	var total int64
	for _, dup := range duplicates {
		total += dup.WastedSize
	}
	return total
}

// blockAlignedSize calculates the size a file occupies when aligned to block boundaries.
// iOS apps store files in 4 KB blocks, so even small files occupy a full block.
func blockAlignedSize(actualSize int64) int64 {
	if actualSize == 0 {
		return 0
	}
	// Round up to nearest block size
	blocks := (actualSize + blockSize - 1) / blockSize
	return blocks * blockSize
}
