package assets

// AssetCatalogInfo contains metadata about an Assets.car file.
type AssetCatalogInfo struct {
	Path          string            `json:"path"`
	TotalSize     int64             `json:"total_size"`
	AssetCount    int               `json:"asset_count"`
	ByType        map[string]int64  `json:"by_type"`
	ByScale       map[string]int64  `json:"by_scale"`
	LargestAssets []AssetInfo       `json:"largest_assets,omitempty"`
}

// AssetInfo contains metadata about a single asset.
type AssetInfo struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Scale string `json:"scale,omitempty"`
	Size  int64  `json:"size"`
}
