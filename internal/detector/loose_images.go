package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

		ext := strings.ToLower(filepath.Ext(path))
		isImage := ext == ".png" || ext == ".jpg" || ext == ".jpeg" ||
			ext == ".gif" || ext == ".webp"

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

// Detect runs the detector and returns a single aggregated optimization
func (d *LooseImagesDetector) Detect(rootPath string) ([]types.Optimization, error) {
	mapper := NewPathMapper(rootPath)
	looseImages, err := DetectLooseImages(rootPath)
	if err != nil || len(looseImages) == 0 {
		return nil, err
	}

	var totalSize int64
	var files []string
	for _, img := range looseImages {
		totalSize += img.Size
		files = append(files, mapper.ToRelative(img.Path))
	}

	if totalSize <= 10*1024 {
		return nil, nil // Skip if too small
	}

	return []types.Optimization{{
		Category:    "loose-images",
		Severity:    "low",
		Title:       fmt.Sprintf("Move %d loose images to asset catalog", len(files)),
		Description: "Images outside asset catalogs don't benefit from app thinning and asset compression",
		Impact:      totalSize / 4, // 25% estimated savings
		Files:       files,
		Action:      "Move images to .xcassets and use asset catalog compilation",
	}}, nil
}
