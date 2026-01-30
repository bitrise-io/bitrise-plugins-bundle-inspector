package detector

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
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

// ImageOptimizationDetector implements the Detector interface
type ImageOptimizationDetector struct{}

// NewImageOptimizationDetector creates a new image optimization detector
func NewImageOptimizationDetector() *ImageOptimizationDetector {
	return &ImageOptimizationDetector{}
}

// checkSipsAvailable verifies that the sips command is available
// sips is built into macOS and required for accurate HEIC conversion measurement
func checkSipsAvailable() error {
	if _, err := exec.LookPath("sips"); err != nil {
		return WrapError("image-optimization", "checking sips availability",
			fmt.Errorf("sips command not found - this tool requires macOS with sips for image optimization detection"))
	}
	return nil
}

// Name returns the detector name
func (d *ImageOptimizationDetector) Name() string {
	return "image-optimization"
}

// hasAlpha checks if a PNG image has an alpha channel (transparency)
// Uses macOS sips command if available, falls back to false if unavailable
func hasAlpha(imagePath string) bool {
	// Check if sips is available
	if _, err := exec.LookPath("sips"); err != nil {
		return false // Assume no alpha if we can't check
	}

	cmd := exec.Command("sips", "-g", "hasAlpha", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// Parse output like "  hasAlpha: yes" or "  hasAlpha: no"
	return strings.Contains(string(output), "hasAlpha: yes")
}

// measureActualHEICConversion converts an image to HEIC and measures real savings
// Supports PNG, JPEG, and WebP source formats
// Returns savings in bytes, or error if conversion fails
func measureActualHEICConversion(imagePath string) (int64, error) {
	// Get original size
	originalInfo, err := os.Stat(imagePath)
	if err != nil {
		return 0, WrapError("image-optimization", "measuring HEIC conversion",
			fmt.Errorf("failed to stat original: %w", err))
	}
	originalSize := originalInfo.Size()

	// Create temp HEIC file
	tmpFile, err := os.CreateTemp("", "heic_conversion_*.heic")
	if err != nil {
		return 0, WrapError("image-optimization", "measuring HEIC conversion",
			fmt.Errorf("failed to create temp file: %w", err))
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Convert to HEIC using sips
	cmd := exec.Command("sips", "-s", "format", "heic", imagePath, "--out", tmpPath)
	if err := cmd.Run(); err != nil {
		return 0, WrapError("image-optimization", "measuring HEIC conversion",
			fmt.Errorf("sips conversion failed: %w", err))
	}

	// Get converted size
	convertedInfo, err := os.Stat(tmpPath)
	if err != nil {
		return 0, WrapError("image-optimization", "measuring HEIC conversion",
			fmt.Errorf("failed to stat converted: %w", err))
	}
	convertedSize := convertedInfo.Size()

	// Calculate savings
	savings := originalSize - convertedSize
	if savings <= 0 {
		return 0, WrapError("image-optimization", "measuring HEIC conversion",
			fmt.Errorf("no savings achieved (HEIC: %d bytes vs original: %d bytes)", convertedSize, originalSize))
	}

	return savings, nil
}

// Detect runs the detector and returns optimizations
func (d *ImageOptimizationDetector) Detect(rootPath string) ([]types.Optimization, error) {
	// Check that sips is available (required for this detector)
	if err := checkSipsAvailable(); err != nil {
		return nil, err
	}

	mapper := util.NewPathMapper(rootPath)
	var optimizations []types.Optimization

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := util.GetLowerExtension(path)

		// Check if this is an image format we can optimize to HEIC
		// PNG, JPEG, and WebP can all benefit from HEIC conversion
		var shouldOptimize bool
		var formatName string
		var minSize int64

		switch ext {
		case ".png":
			shouldOptimize = true
			formatName = "PNG"
			minSize = 5 * 1024 // 5 KB threshold for PNGs
		case ".jpg", ".jpeg":
			shouldOptimize = true
			formatName = "JPEG"
			minSize = 10 * 1024 // 10 KB threshold for JPEGs
		case ".webp":
			shouldOptimize = true
			formatName = "WebP"
			minSize = 10 * 1024 // 10 KB threshold for WebP
		}

		if !shouldOptimize || info.Size() < minSize {
			return nil
		}

		// Measure actual HEIC conversion savings
		savings, err := measureActualHEICConversion(path)
		if err != nil {
			// Skip this image if conversion fails
			return nil
		}

		// Check for transparency (PNG only, others don't need the check)
		alphaNote := ""
		if ext == ".png" && hasAlpha(path) {
			alphaNote = " (has transparency - supported in iOS 11+)"
		}

		// Build description based on format
		description := fmt.Sprintf("%s can be converted to HEIC format for better compression. "+
			"Measured savings: %s%s", formatName, util.FormatBytes(savings), alphaNote)

		optimizations = append(optimizations, types.Optimization{
			Category:    "image-optimization",
			Severity:    "medium",
			Title:       fmt.Sprintf("Convert %s to HEIC", filepath.Base(path)),
			Description: description,
			Impact:      savings,
			Files:       []string{mapper.ToRelative(path)},
			Action:      "Convert to HEIC format using Xcode Asset Catalog (Image Set with Preserve Vector Data disabled)",
		})

		return nil
	})

	if err != nil {
		return nil, WrapError("image-optimization", "detecting optimizations", err)
	}

	return optimizations, nil
}
