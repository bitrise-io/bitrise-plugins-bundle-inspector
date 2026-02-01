//go:build darwin

package assets

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// assetutilEntry represents a single entry from assetutil JSON output.
// The first entry is metadata, subsequent entries are assets.
type assetutilEntry struct {
	// Metadata fields (first entry)
	AssetStorageVersion string `json:"AssetStorageVersion,omitempty"`
	Platform            string `json:"Platform,omitempty"`

	// Asset fields
	AssetType     string  `json:"AssetType,omitempty"`
	Name          string  `json:"Name,omitempty"`
	RenditionName string  `json:"RenditionName,omitempty"`
	Scale         float64 `json:"Scale,omitempty"`
	SizeOnDisk    int64   `json:"SizeOnDisk,omitempty"`
	Idiom         string  `json:"Idiom,omitempty"`
	Compression   string  `json:"Compression,omitempty"`
	PixelWidth    int     `json:"PixelWidth,omitempty"`
	PixelHeight   int     `json:"PixelHeight,omitempty"`
	SHA1Digest    string  `json:"SHA1Digest,omitempty"`
}

// ParseAssetCatalog extracts metadata from an Assets.car file using assetutil.
func ParseAssetCatalog(carPath string) (*AssetCatalogInfo, error) {
	// Get file size
	fileInfo, err := os.Stat(carPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat Assets.car: %w", err)
	}

	catalog := &AssetCatalogInfo{
		Path:       filepath.Base(carPath),
		TotalSize:  fileInfo.Size(),
		AssetCount: 0,
		ByType:     make(map[string]int64),
		ByScale:    make(map[string]int64),
	}

	// Run assetutil to get asset information
	assets, err := runAssetutil(carPath)
	if err != nil {
		// Graceful fallback: return basic info from file stat
		return catalog, nil
	}

	catalog.Assets = assets
	catalog.AssetCount = len(assets)

	// Categorize assets
	byType, byScale := CategorizeAssets(assets)
	catalog.ByType = byType
	catalog.ByScale = byScale

	// Find largest assets (top 10)
	catalog.LargestAssets = findLargestAssets(assets, 10)

	return catalog, nil
}

// runAssetutil executes assetutil and parses the JSON output.
func runAssetutil(carPath string) ([]AssetInfo, error) {
	cmd := exec.Command("assetutil", "-I", carPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("assetutil failed: %w", err)
	}

	var entries []assetutilEntry
	if err := json.Unmarshal(output, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse assetutil output: %w", err)
	}

	var assets []AssetInfo
	for _, entry := range entries {
		// Skip metadata entry (first entry has AssetStorageVersion)
		if entry.AssetStorageVersion != "" {
			continue
		}

		// Skip entries without a name
		if entry.Name == "" {
			continue
		}

		asset := AssetInfo{
			Name:          entry.Name,
			RenditionName: entry.RenditionName,
			Type:          normalizeAssetType(entry.AssetType),
			Scale:         formatScale(entry.Scale),
			Size:          entry.SizeOnDisk,
			Idiom:         strings.ToLower(entry.Idiom),
			Compression:   entry.Compression,
			PixelWidth:    entry.PixelWidth,
			PixelHeight:   entry.PixelHeight,
			SHA1Digest:    entry.SHA1Digest,
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

// normalizeAssetType converts assetutil asset types to consistent lowercase names.
func normalizeAssetType(assetType string) string {
	lower := strings.ToLower(assetType)

	switch lower {
	case "image":
		return "image"
	case "icon image", "icon":
		return "icon"
	case "vector":
		return "vector"
	case "color", "namedcolor", "named color":
		return "color"
	case "data":
		return "data"
	case "imageset", "image set":
		return "imageset"
	case "iconset", "icon set", "appicon", "app icon":
		return "icon"
	case "multisize image catalog entry", "multisize image":
		return "image"
	case "packed atlas image":
		return "image"
	default:
		if assetType == "" {
			return "unknown"
		}
		// Handle compound types like "Packed Atlas Image"
		if strings.Contains(lower, "image") {
			return "image"
		}
		if strings.Contains(lower, "icon") {
			return "icon"
		}
		if strings.Contains(lower, "color") {
			return "color"
		}
		return strings.ToLower(strings.ReplaceAll(assetType, " ", "_"))
	}
}

// formatScale converts numeric scale to string format (e.g., 2 -> "2x").
func formatScale(scale float64) string {
	if scale <= 0 {
		return ""
	}
	if scale == float64(int(scale)) {
		return fmt.Sprintf("%dx", int(scale))
	}
	return fmt.Sprintf("%.1fx", scale)
}

// CategorizeAssets groups assets by type and scale.
func CategorizeAssets(assets []AssetInfo) (byType, byScale map[string]int64) {
	byType = make(map[string]int64)
	byScale = make(map[string]int64)

	for _, asset := range assets {
		byType[asset.Type] += asset.Size
		if asset.Scale != "" {
			byScale[asset.Scale] += asset.Size
		}
	}

	return byType, byScale
}

// findLargestAssets returns the N largest assets.
func findLargestAssets(assets []AssetInfo, n int) []AssetInfo {
	if len(assets) == 0 {
		return nil
	}

	// Sort by size descending
	sorted := make([]AssetInfo, len(assets))
	copy(sorted, assets)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Size > sorted[j].Size
	})

	// Return top N
	if len(sorted) > n {
		sorted = sorted[:n]
	}

	return sorted
}

// BuildVirtualAssetName creates a filename for virtual asset display.
// Format: <Name>~<idiom>@<scale>x.<extension>
func BuildVirtualAssetName(asset AssetInfo) string {
	name := asset.Name

	// Add idiom if present
	if asset.Idiom != "" && asset.Idiom != "universal" {
		name += "~" + asset.Idiom
	}

	// Add scale if present and not 1x
	if asset.Scale != "" && asset.Scale != "1x" {
		name += "@" + asset.Scale
	}

	// Add extension based on type
	switch asset.Type {
	case "image", "icon", "imageset":
		name += ".png"
	case "vector":
		name += ".pdf"
	case "color":
		// No extension for colors
	case "data":
		name += ".data"
	default:
		if asset.Type != "" && asset.Type != "unknown" {
			name += "." + asset.Type
		}
	}

	return name
}

// ExpandAssetsAsChildren creates virtual FileNode children from asset catalog assets.
// The virtual nodes represent the individual assets within the .car file.
func ExpandAssetsAsChildren(catalog *AssetCatalogInfo, carRelativePath string) []*types.FileNode {
	if catalog == nil || len(catalog.Assets) == 0 {
		return nil
	}

	children := make([]*types.FileNode, 0, len(catalog.Assets))
	for _, asset := range catalog.Assets {
		virtualName := BuildVirtualAssetName(asset)
		virtualPath := filepath.Join(carRelativePath, virtualName)

		node := &types.FileNode{
			Path:       virtualPath,
			Name:       virtualName,
			Size:       asset.Size,
			IsDir:      false,
			IsVirtual:  true,
			SourceFile: carRelativePath,
		}
		children = append(children, node)
	}

	// Sort by size descending for better treemap visualization
	sort.Slice(children, func(i, j int) bool {
		return children[i].Size > children[j].Size
	})

	return children
}
