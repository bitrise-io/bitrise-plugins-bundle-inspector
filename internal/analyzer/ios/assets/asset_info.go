package assets

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
