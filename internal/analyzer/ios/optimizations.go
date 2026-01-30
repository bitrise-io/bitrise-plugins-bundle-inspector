package ios

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/assets"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// GenerateUnusedFrameworkOptimizations creates optimization suggestions for unused frameworks
func GenerateUnusedFrameworkOptimizations(unusedFrameworks []string, frameworks []*FrameworkInfo) []types.Optimization {
	var optimizations []types.Optimization

	for _, fwPath := range unusedFrameworks {
		// Find framework info to get size
		var fwSize int64
		var fwName string
		for _, fw := range frameworks {
			if strings.Contains(fwPath, fw.Name) {
				fwSize = fw.Size
				fwName = fw.Name
				break
			}
		}

		if fwName != "" {
			optimizations = append(optimizations, types.Optimization{
				Category:    "frameworks",
				Severity:    "medium",
				Title:       fmt.Sprintf("Unused framework: %s", fwName),
				Description: "Framework is not linked by main binary or other frameworks",
				Action:      "Remove framework to reduce app size",
				Files:       []string{fwPath},
				Impact:      fwSize,
			})
		}
	}

	return optimizations
}

// GenerateLargeAssetOptimizations creates optimization suggestions for oversized assets
func GenerateLargeAssetOptimizations(assetCatalogs []*assets.AssetCatalogInfo) []types.Optimization {
	var optimizations []types.Optimization

	for _, catalog := range assetCatalogs {
		for _, asset := range catalog.LargestAssets {
			if asset.Size > 1*1024*1024 { // >1MB
				optimizations = append(optimizations, types.Optimization{
					Category:    "assets",
					Severity:    "low",
					Title:       fmt.Sprintf("Large asset: %s", asset.Name),
					Description: fmt.Sprintf("Asset is %s", util.FormatBytes(asset.Size)),
					Files:       []string{asset.Name},
					Impact:      asset.Size,
					Action:      "Consider compressing or resizing asset",
				})
			}
		}
	}

	return optimizations
}
