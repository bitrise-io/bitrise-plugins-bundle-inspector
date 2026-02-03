// Package types provides public API types for the bundle inspector.
package types

import "time"

// ArtifactType represents the type of mobile artifact being analyzed.
type ArtifactType string

const (
	ArtifactTypeIPA       ArtifactType = "ipa"
	ArtifactTypeAPK       ArtifactType = "apk"
	ArtifactTypeAAB       ArtifactType = "aab"
	ArtifactTypeXCArchive ArtifactType = "xcarchive"
	ArtifactTypeApp       ArtifactType = "app"
)

// ArtifactInfo contains basic information about the analyzed artifact.
type ArtifactInfo struct {
	Path             string       `json:"path"`
	Type             ArtifactType `json:"type"`
	Size             int64        `json:"size"`
	UncompressedSize int64        `json:"uncompressed_size,omitempty"`
	AnalyzedAt       time.Time    `json:"analyzed_at"`
	IconData         string       `json:"icon_data,omitempty"`  // Base64-encoded icon (data URI format)
	AppName          string       `json:"app_name,omitempty"`   // App display name
	BundleID         string       `json:"bundle_id,omitempty"`  // Bundle/package identifier
	Version          string       `json:"version,omitempty"`    // App version
}

// SizeBreakdown provides a categorized breakdown of artifact size.
type SizeBreakdown struct {
	Executable  int64            `json:"executable"`
	Frameworks  int64            `json:"frameworks"`
	Resources   int64            `json:"resources"`
	Assets      int64            `json:"assets"`
	Libraries   int64            `json:"libraries"`
	DEX         int64            `json:"dex,omitempty"`
	Other       int64            `json:"other"`
	ByCategory  map[string]int64 `json:"by_category,omitempty"`
	ByExtension map[string]int64 `json:"by_extension,omitempty"`
}

// FileNode represents a file or directory in the artifact tree.
type FileNode struct {
	Path       string                 `json:"path"`
	Name       string                 `json:"name"`
	Size       int64                  `json:"size"`
	IsDir      bool                   `json:"is_dir"`
	Children   []*FileNode            `json:"children,omitempty"`
	IsVirtual  bool                   `json:"is_virtual,omitempty"`  // True for assets expanded from .car files or DEX classes
	SourceFile string                 `json:"source_file,omitempty"` // Parent .car file path or DEX file for virtual nodes
	Metadata   map[string]interface{} `json:"metadata,omitempty"`    // Additional metadata (e.g., for DEX classes)
}

// DuplicateSet represents a group of duplicate files.
type DuplicateSet struct {
	Hash       string   `json:"hash"`
	Size       int64    `json:"size"`
	Count      int      `json:"count"`
	Files      []string `json:"files"`
	WastedSize int64    `json:"wasted_size"` // (count - 1) * size
}

// Optimization represents a potential size optimization.
type Optimization struct {
	Category    string   `json:"category"`    // "duplicates", "compression", "architecture", etc.
	Severity    string   `json:"severity"`    // "high", "medium", "low"
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      int64    `json:"impact"`       // Estimated savings in bytes
	Files       []string `json:"files"`
	Action      string   `json:"action"`       // Suggested action
}

// Report contains the complete analysis results.
type Report struct {
	ArtifactInfo   ArtifactInfo           `json:"artifact_info"`
	SizeBreakdown  SizeBreakdown          `json:"size_breakdown"`
	FileTree       []*FileNode            `json:"file_tree"`
	Duplicates     []DuplicateSet         `json:"duplicates,omitempty"`
	Optimizations  []Optimization         `json:"optimizations,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	LargestFiles   []FileNode             `json:"largest_files,omitempty"`
	TotalSavings   int64                  `json:"total_savings,omitempty"`
}

// BinaryInfo contains parsed Mach-O metadata (iOS binaries).
type BinaryInfo struct {
	Architecture     string   `json:"architecture"`
	Architectures    []string `json:"architectures"`
	Type             string   `json:"type"`
	CodeSize         int64    `json:"code_size"`
	DataSize         int64    `json:"data_size"`
	LinkedLibraries  []string `json:"linked_libraries"`
	RPaths           []string `json:"rpaths,omitempty"`
	HasDebugSymbols  bool     `json:"has_debug_symbols"`
	DebugSymbolsSize int64    `json:"debug_symbols_size,omitempty"`
}

// FrameworkInfo contains metadata about an iOS framework.
type FrameworkInfo struct {
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	Version      string      `json:"version,omitempty"`
	Size         int64       `json:"size"`
	BinaryInfo   *BinaryInfo `json:"binary_info,omitempty"`
	Dependencies []string    `json:"dependencies,omitempty"`
}

// AssetCatalogInfo contains metadata about an Assets.car file.
type AssetCatalogInfo struct {
	Path          string            `json:"path"`
	TotalSize     int64             `json:"total_size"`
	AssetCount    int               `json:"asset_count"`
	ByType        map[string]int64  `json:"by_type"`
	ByScale       map[string]int64  `json:"by_scale"`
	LargestAssets []AssetInfo       `json:"largest_assets,omitempty"`
	Assets        []AssetInfo       `json:"assets,omitempty"` // All assets in the catalog
}

// AssetInfo contains metadata about a single asset.
type AssetInfo struct {
	Name          string `json:"name"`
	RenditionName string `json:"rendition_name,omitempty"`
	Type          string `json:"type"`
	Scale         string `json:"scale,omitempty"`
	Size          int64  `json:"size"`
	Idiom         string `json:"idiom,omitempty"`
	Compression   string `json:"compression,omitempty"`
	PixelWidth    int    `json:"pixel_width,omitempty"`
	PixelHeight   int    `json:"pixel_height,omitempty"`
	SHA1Digest    string `json:"sha1_digest,omitempty"`
}

// DexInfo contains parsed DEX file information.
type DexInfo struct {
	SourceFile       string                 `json:"source_file"`
	Classes          []DexClass             `json:"classes"`
	TotalPrivateSize int64                  `json:"total_private_size"`
	TotalFileSize    int64                  `json:"total_file_size"`
	IsObfuscated     bool                   `json:"is_obfuscated"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// DexClass represents a single class in a DEX file.
type DexClass struct {
	ClassName   string                 `json:"class_name"`
	PackageName string                 `json:"package_name"`
	PrivateSize int64                  `json:"private_size"`
	MethodCount int                    `json:"method_count"`
	FieldCount  int                    `json:"field_count"`
	SourceDEX   string                 `json:"source_dex"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MergedDEXInfo contains information from multiple DEX files.
type MergedDEXInfo struct {
	Classes          []DexClass `json:"classes"`
	TotalPrivateSize int64      `json:"total_private_size"`
	TotalFileSize    int64      `json:"total_file_size"`
	DEXFileCount     int        `json:"dex_file_count"`
}
