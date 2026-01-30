package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// LooseImage represents an image outside asset catalogs
type LooseImage struct {
	Path           string
	Size           int64
	InAssetCatalog bool
}

// DetectLooseImages finds images outside asset catalogs
func DetectLooseImages(rootPath string) ([]LooseImage, error) {
	var looseImages []LooseImage
	var assetCatalogPaths = make(map[string]bool)

	// First pass: find all asset catalogs
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasSuffix(path, ".xcassets") {
			assetCatalogPaths[path] = true
		}
		return nil
	})

	// Second pass: find images and check if they're in catalogs
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		isImage := util.HasExtension(path, ".png", ".jpg", ".jpeg", ".gif", ".webp")

		if !isImage {
			return nil
		}

		// Check if image is inside an asset catalog
		inCatalog := false
		for catalogPath := range assetCatalogPaths {
			if strings.HasPrefix(path, catalogPath) {
				inCatalog = true
				break
			}
		}

		// Also skip if in Assets.car (compiled asset catalog)
		if strings.Contains(path, "Assets.car") {
			inCatalog = true
		}

		if !inCatalog {
			looseImages = append(looseImages, LooseImage{
				Path:           path,
				Size:           info.Size(),
				InAssetCatalog: false,
			})
		}

		return nil
	})

	return looseImages, err
}

// LooseImagesDetector implements the Detector interface
type LooseImagesDetector struct{}

// NewLooseImagesDetector creates a new loose images detector
func NewLooseImagesDetector() *LooseImagesDetector {
	return &LooseImagesDetector{}
}

// Name returns the detector name
func (d *LooseImagesDetector) Name() string {
	return "loose-images"
}

// imagePattern represents a group of related images
type imagePattern struct {
	baseName  string
	images    []LooseImage
	patternType string // "retina-variants", "multi-location", or "other"
}

// extractBaseName removes retina scale suffixes (@1x, @2x, @3x) from filename
func extractBaseName(filename string) string {
	// Match pattern like "Icon@2x.png" → "Icon"
	re := regexp.MustCompile(`@[123]x`)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	baseName := re.ReplaceAllString(nameWithoutExt, "")
	return baseName
}

// detectPatterns groups loose images by base name and identifies redundancy patterns
func detectPatterns(looseImages []LooseImage) []imagePattern {
	// Group by base filename (without path and scale suffix)
	grouped := make(map[string][]LooseImage)

	for _, img := range looseImages {
		filename := filepath.Base(img.Path)
		baseName := extractBaseName(filename)
		ext := filepath.Ext(filename)
		key := baseName + ext // Group by base name + extension

		grouped[key] = append(grouped[key], img)
	}

	var patterns []imagePattern

	for baseName, images := range grouped {
		if len(images) < 2 {
			continue // Skip single images, not a pattern
		}

		// Detect pattern type
		patternType := detectPatternType(images)
		if patternType == "" {
			continue // No detectable redundancy pattern
		}

		patterns = append(patterns, imagePattern{
			baseName:    baseName,
			images:      images,
			patternType: patternType,
		})
	}

	return patterns
}

// detectPatternType identifies what kind of redundancy exists
func detectPatternType(images []LooseImage) string {
	// Check for retina variants (@1x, @2x, @3x)
	hasRetinaVariants := false
	for _, img := range images {
		filename := filepath.Base(img.Path)
		if strings.Contains(filename, "@2x") || strings.Contains(filename, "@3x") || strings.Contains(filename, "@1x") {
			hasRetinaVariants = true
			break
		}
	}

	if hasRetinaVariants {
		return "retina-variants"
	}

	// Check for same filename in multiple locations with different sizes
	sizeMap := make(map[int64]bool)
	for _, img := range images {
		sizeMap[img.Size] = true
	}

	if len(sizeMap) > 1 {
		return "multi-location"
	}

	return "" // No detectable pattern
}

// calculatePatternSavings computes redundancy savings for a pattern
// Uses disk usage (4 KB blocks) instead of file sizes for accurate savings
func calculatePatternSavings(pattern imagePattern) int64 {
	if len(pattern.images) < 2 {
		return 0
	}

	switch pattern.patternType {
	case "retina-variants":
		// For retina variants, we keep the largest and generate others
		// Savings = sum of disk usage of all smaller variants
		var totalDiskUsage int64
		var maxDiskUsage int64

		for _, img := range pattern.images {
			diskUsage := util.CalculateDiskUsage(img.Size)
			totalDiskUsage += diskUsage
			if diskUsage > maxDiskUsage {
				maxDiskUsage = diskUsage
			}
		}

		// Save everything except the largest variant
		return totalDiskUsage - maxDiskUsage

	case "multi-location":
		// For multiple locations with different sizes, keep the largest
		// Savings = sum of disk usage of smaller files
		var totalDiskUsage int64
		var maxDiskUsage int64

		for _, img := range pattern.images {
			diskUsage := util.CalculateDiskUsage(img.Size)
			totalDiskUsage += diskUsage
			if diskUsage > maxDiskUsage {
				maxDiskUsage = diskUsage
			}
		}

		// Save everything except the largest
		return totalDiskUsage - maxDiskUsage

	default:
		return 0
	}
}

// createPatternOptimization generates an optimization from a detected pattern
func createPatternOptimization(pattern imagePattern, mapper *util.PathMapper) *types.Optimization {
	savings := calculatePatternSavings(pattern)
	if savings <= 0 {
		return nil
	}

	// Collect and sort file paths
	var files []string
	for _, img := range pattern.images {
		files = append(files, mapper.ToRelative(img.Path))
	}
	sort.Strings(files)

	// Generate title and description based on pattern type
	var title, description string
	switch pattern.patternType {
	case "retina-variants":
		title = fmt.Sprintf("Consolidate %s into asset catalog", pattern.baseName)
		description = fmt.Sprintf(
			"Found %d retina scale variants that can be auto-generated from a single @3x image. "+
				"Asset catalogs automatically generate @1x and @2x from @3x, eliminating manual variants.",
			len(pattern.images))

	case "multi-location":
		title = fmt.Sprintf("Consolidate variants of %s into asset catalog", pattern.baseName)
		description = fmt.Sprintf(
			"Same image appears in %d locations with different sizes (likely manual resizing). "+
				"Use asset catalogs to keep the largest version and auto-generate other sizes as needed.",
			len(pattern.images))
	}

	return &types.Optimization{
		Category:    "loose-images",
		Severity:    "low",
		Title:       title,
		Description: description,
		Impact:      savings,
		Files:       files,
		Action:      "Move images to .xcassets and enable app thinning with automatic scale generation",
	}
}

// Detect runs the detector and returns pattern-based optimizations
func (d *LooseImagesDetector) Detect(rootPath string) ([]types.Optimization, error) {
	mapper := util.NewPathMapper(rootPath)

	// Detect all loose images
	looseImages, err := DetectLooseImages(rootPath)
	if err != nil {
		return nil, WrapError("loose-images", "detecting loose images", err)
	}
	if len(looseImages) == 0 {
		return nil, nil
	}

	// Detect patterns (retina variants, multi-location duplicates)
	patterns := detectPatterns(looseImages)

	// Generate optimizations from patterns
	var optimizations []types.Optimization
	for _, pattern := range patterns {
		if opt := createPatternOptimization(pattern, mapper); opt != nil {
			optimizations = append(optimizations, *opt)
		}
	}

	// Sort by impact (highest first)
	sort.Slice(optimizations, func(i, j int) bool {
		return optimizations[i].Impact > optimizations[j].Impact
	})

	return optimizations, nil
}
