package detector

import (
	"os"
	"path/filepath"
	"strings"
)

// ImageOptimization represents an image that could be optimized
type ImageOptimization struct {
	Path              string
	CurrentFormat     string
	CurrentSize       int64
	RecommendedFormat string
	EstimatedSavings  int64
}

// DetectImageOptimizations scans for images that could be optimized
func DetectImageOptimizations(rootPath string) ([]ImageOptimization, error) {
	var optimizations []ImageOptimization

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		// Check PNG optimization opportunity
		if ext == ".png" && info.Size() > 10*1024 { // > 10 KB
			// Estimate 20-30% savings from PNG optimization
			savings := int64(float64(info.Size()) * 0.25)

			optimizations = append(optimizations, ImageOptimization{
				Path:              path,
				CurrentFormat:     "PNG",
				CurrentSize:       info.Size(),
				RecommendedFormat: "Optimized PNG or HEIC",
				EstimatedSavings:  savings,
			})
		}

		// Check JPEG optimization opportunity
		if (ext == ".jpg" || ext == ".jpeg") && info.Size() > 50*1024 { // > 50 KB
			// Large JPEGs could benefit from HEIC conversion
			savings := int64(float64(info.Size()) * 0.4) // 40% savings typical for HEIC

			optimizations = append(optimizations, ImageOptimization{
				Path:              path,
				CurrentFormat:     "JPEG",
				CurrentSize:       info.Size(),
				RecommendedFormat: "HEIC",
				EstimatedSavings:  savings,
			})
		}

		return nil
	})

	return optimizations, err
}
