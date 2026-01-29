// Package detector provides file analysis and detection capabilities.
package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// BloatDetector identifies large files and potential optimization opportunities.
type BloatDetector struct {
	threshold int64 // Size threshold in bytes
}

// NewBloatDetector creates a new bloat detector with the given threshold.
func NewBloatDetector(threshold int64) *BloatDetector {
	return &BloatDetector{
		threshold: threshold,
	}
}

// DetectLargeFiles finds files exceeding the size threshold.
func (d *BloatDetector) DetectLargeFiles(rootPath string) ([]types.FileNode, error) {
	var largeFiles []types.FileNode

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			return nil
		}

		if info.Size() >= d.threshold {
			// Get relative path from root
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				relPath = path
			}

			largeFiles = append(largeFiles, types.FileNode{
				Path:  relPath,
				Name:  filepath.Base(path),
				Size:  info.Size(),
				IsDir: false,
			})
		}

		return nil
	})

	return largeFiles, err
}

// DetectCompressionOpportunities identifies uncompressed image files.
func (d *BloatDetector) DetectCompressionOpportunities(files []types.FileNode) []types.FileNode {
	var opportunities []types.FileNode

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Path))

		// Check for uncompressed image formats
		if ext == ".bmp" || ext == ".tiff" || ext == ".tif" {
			opportunities = append(opportunities, file)
		}

		// Large PNGs might benefit from compression
		if ext == ".png" && file.Size > 1024*1024 { // > 1MB
			opportunities = append(opportunities, file)
		}
	}

	return opportunities
}

// AnalyzeResourceTypes categorizes resources by type and size.
func (d *BloatDetector) AnalyzeResourceTypes(rootPath string) (map[string]int64, error) {
	resourceSizes := make(map[string]int64)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(filepath.Base(path)))
		if ext != "" {
			resourceSizes[ext] += info.Size()
		}

		return nil
	})

	return resourceSizes, err
}
