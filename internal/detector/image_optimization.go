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
type ImageOptimizationDetector struct {
	platform Platform
}

// NewImageOptimizationDetector creates a new image optimization detector for the given platform
func NewImageOptimizationDetector(platform Platform) *ImageOptimizationDetector {
	return &ImageOptimizationDetector{platform: platform}
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

// measureActualWebPConversion converts an image to WebP using cwebp and measures real savings.
// Returns savings in bytes, or error if cwebp is not available or conversion fails.
func measureActualWebPConversion(imagePath string) (int64, error) {
	if _, err := exec.LookPath("cwebp"); err != nil {
		return 0, fmt.Errorf("cwebp not available")
	}

	originalInfo, err := os.Stat(imagePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat original: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "webp_conversion_*.webp")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command("cwebp", "-q", "80", imagePath, "-o", tmpPath)
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("cwebp conversion failed: %w", err)
	}

	convertedInfo, err := os.Stat(tmpPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat converted: %w", err)
	}

	savings := originalInfo.Size() - convertedInfo.Size()
	if savings <= 0 {
		return 0, fmt.Errorf("no savings achieved (WebP: %d bytes vs original: %d bytes)", convertedInfo.Size(), originalInfo.Size())
	}

	return savings, nil
}

// estimateWebPSavings estimates savings from converting to WebP format using
// conservative compression ratios based on published benchmarks.
func estimateWebPSavings(imagePath string) (int64, error) {
	info, err := os.Stat(imagePath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	originalSize := info.Size()

	// Conservative estimate: ~25% savings for both PNG and JPEG
	// Based on Google's published WebP compression benchmarks
	savings := int64(float64(originalSize) * 0.25)
	if savings <= 0 {
		return 0, fmt.Errorf("no estimated savings")
	}

	return savings, nil
}

// targetFormat returns the recommended image format name for this platform
func (d *ImageOptimizationDetector) targetFormat() string {
	if d.platform == PlatformAndroid {
		return "WebP"
	}
	return "HEIC"
}

// shouldOptimizeImage checks if a file extension is eligible for optimization on this platform
func (d *ImageOptimizationDetector) shouldOptimizeImage(ext string) (shouldOptimize bool, formatName string, minSize int64) {
	switch ext {
	case ".png":
		return true, "PNG", 5 * 1024
	case ".jpg", ".jpeg":
		return true, "JPEG", 10 * 1024
	case ".webp":
		// WebP files are already in the target format for Android
		if d.platform == PlatformAndroid {
			return false, "", 0
		}
		return true, "WebP", 10 * 1024
	}
	return false, "", 0
}

// measureSavings measures or estimates savings for converting an image to the target format
func (d *ImageOptimizationDetector) measureSavings(imagePath string) (int64, error) {
	if d.platform == PlatformAndroid {
		// Try actual cwebp measurement first, fall back to estimation
		if savings, err := measureActualWebPConversion(imagePath); err == nil {
			return savings, nil
		}
		return estimateWebPSavings(imagePath)
	}
	return measureActualHEICConversion(imagePath)
}

// buildRecommendation creates platform-appropriate description and action text
func (d *ImageOptimizationDetector) buildRecommendation(formatName, imagePath, ext string, savings int64) (description, action string) {
	if d.platform == PlatformAndroid {
		description = fmt.Sprintf("%s can be converted to WebP format for better compression. "+
			"Estimated savings: %s", formatName, util.FormatBytes(savings))
		action = "Convert to WebP format (supported since Android 4.0 for lossy, Android 4.3 for lossless and transparency)"
		return
	}

	alphaNote := ""
	if ext == ".png" && hasAlpha(imagePath) {
		alphaNote = " (has transparency - supported in iOS 11+)"
	}
	description = fmt.Sprintf("%s can be converted to HEIC format for better compression. "+
		"Measured savings: %s%s", formatName, util.FormatBytes(savings), alphaNote)
	action = "Convert to HEIC format using Xcode Asset Catalog (Image Set with Preserve Vector Data disabled)"
	return
}

// Detect runs the detector and returns optimizations
func (d *ImageOptimizationDetector) Detect(rootPath string) ([]types.Optimization, error) {
	// iOS requires sips for actual HEIC conversion measurement
	if d.platform == PlatformIOS {
		if err := checkSipsAvailable(); err != nil {
			return nil, err
		}
	}

	mapper := util.NewPathMapper(rootPath)
	var optimizations []types.Optimization

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := util.GetLowerExtension(path)

		shouldOptimize, formatName, minSize := d.shouldOptimizeImage(ext)
		if !shouldOptimize || info.Size() < minSize {
			return nil
		}

		savings, err := d.measureSavings(path)
		if err != nil {
			// Skip this image if measurement/estimation fails
			return nil
		}

		description, action := d.buildRecommendation(formatName, path, ext, savings)

		optimizations = append(optimizations, types.Optimization{
			Category:    "image-optimization",
			Severity:    "medium",
			Title:       fmt.Sprintf("Convert %s to %s", filepath.Base(path), d.targetFormat()),
			Description: description,
			Impact:      savings,
			Files:       []string{mapper.ToRelative(path)},
			Action:      action,
		})

		return nil
	})

	if err != nil {
		return nil, WrapError("image-optimization", "detecting optimizations", err)
	}

	return optimizations, nil
}
