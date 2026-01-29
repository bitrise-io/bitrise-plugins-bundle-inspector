package detector

import (
	"os"
	"path/filepath"
	"strings"
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
