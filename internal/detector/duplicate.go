// Package detector provides file analysis and detection capabilities.
package detector

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
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

		dup := types.DuplicateSet{
			Hash:       hash,
			Size:       info.Size(),
			Count:      len(files),
			Files:      files,
			WastedSize: (int64(len(files)) - 1) * info.Size(),
		}
		duplicates = append(duplicates, dup)
	}

	return duplicates, nil
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
