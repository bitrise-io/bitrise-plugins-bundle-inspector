package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
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
		if ext == ".png" && info.Size() > 5*1024 { // > 5 KB
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

// ImageOptimizationDetector implements the Detector interface
type ImageOptimizationDetector struct{}

// NewImageOptimizationDetector creates a new image optimization detector
func NewImageOptimizationDetector() *ImageOptimizationDetector {
	return &ImageOptimizationDetector{}
}

// Name returns the detector name
func (d *ImageOptimizationDetector) Name() string {
	return "image-optimization"
}

// Detect runs the detector and returns optimizations
func (d *ImageOptimizationDetector) Detect(rootPath string) ([]types.Optimization, error) {
	mapper := NewPathMapper(rootPath)
	var optimizations []types.Optimization

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		// Check PNG optimization opportunity
		if ext == ".png" && info.Size() > 10*1024 {
			savings := int64(float64(info.Size()) * 0.25)
			optimizations = append(optimizations, types.Optimization{
				Category:    "image-optimization",
				Severity:    "medium",
				Title:       fmt.Sprintf("Optimize %s", filepath.Base(path)),
				Description: "Convert PNG to Optimized PNG or HEIC for better compression",
				Impact:      savings,
				Files:       []string{mapper.ToRelative(path)},
				Action:      "Convert to Optimized PNG or HEIC or optimize with image compression tools",
			})
		}

		// Check JPEG optimization opportunity
		if (ext == ".jpg" || ext == ".jpeg") && info.Size() > 50*1024 {
			savings := int64(float64(info.Size()) * 0.4)
			optimizations = append(optimizations, types.Optimization{
				Category:    "image-optimization",
				Severity:    "medium",
				Title:       fmt.Sprintf("Optimize %s", filepath.Base(path)),
				Description: "Convert JPEG to HEIC for better compression",
				Impact:      savings,
				Files:       []string{mapper.ToRelative(path)},
				Action:      "Convert to HEIC or optimize with image compression tools",
			})
		}

		return nil
	})

	return optimizations, err
}
