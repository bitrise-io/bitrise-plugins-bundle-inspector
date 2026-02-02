package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// HTMLFormatter formats reports as interactive HTML with treemap visualization
type HTMLFormatter struct {
	Title          string
	IncludeTreemap bool
	IncludeCharts  bool
	Theme          string // "light" or "dark"
}

// NewHTMLFormatter creates a new HTML formatter
func NewHTMLFormatter() *HTMLFormatter {
	return &HTMLFormatter{
		Title:          "Bundle Analysis Report",
		IncludeTreemap: true,
		IncludeCharts:  true,
		Theme:          "light",
	}
}

// Format writes the report in HTML format to the writer
func (f *HTMLFormatter) Format(w io.Writer, report *types.Report) error {
	// Prepare data for JavaScript
	data := f.prepareTemplateData(report)

	// Execute template
	return f.executeTemplate(w, data)
}

// templateData holds all data needed for the HTML template
type templateData struct {
	Title              string
	AppName            string
	BundleID           string
	Platform           string
	Version            string
	Branch             string
	CommitSHA          string
	ArtifactName       string
	ArtifactType       string
	TotalSize          string
	UncompressedSize   string
	CompressionRatio   string
	TotalSavings       string
	SavingsPercentage  string
	Timestamp          string
	DataJSON           template.JS
	NodeCount          int
	PerformanceWarning bool
}

// prepareTemplateData converts the report into template-ready data
func (f *HTMLFormatter) prepareTemplateData(report *types.Report) templateData {
	// Extract artifact name
	artifactName := report.ArtifactInfo.Path
	if idx := strings.LastIndex(artifactName, "/"); idx >= 0 {
		artifactName = artifactName[idx+1:]
	}

	// Calculate sizes
	uncompressedSize := calculateUncompressedSize(&report.SizeBreakdown)
	downloadSize := report.ArtifactInfo.Size

	// Calculate compression ratio
	compressionRatio := "N/A"
	if uncompressedSize > 0 && downloadSize > 0 {
		ratio := (1.0 - float64(downloadSize)/float64(uncompressedSize)) * 100
		compressionRatio = fmt.Sprintf("%.1f%%", ratio)
	}

	// Calculate savings percentage
	savingsPercentage := "0.0%"
	if uncompressedSize > 0 {
		savingsPercentage = fmt.Sprintf("%.1f%%", float64(report.TotalSavings)/float64(uncompressedSize)*100)
	}

	// Count nodes for performance warning
	nodeCount := f.calculateNodeCount(report.FileTree)
	performanceWarning := nodeCount > 10000

	// Prepare data structure for JavaScript
	jsData := f.prepareJSData(report)
	dataJSON, err := json.Marshal(jsData)
	if err != nil {
		dataJSON = []byte("{}")
	}

	// Extract metadata fields with safe type assertions
	appName := ""
	bundleID := ""
	platform := ""
	version := ""
	branch := ""
	commitSHA := ""

	if report.Metadata != nil {
		if v, ok := report.Metadata["app_name"].(string); ok {
			appName = v
		}
		if v, ok := report.Metadata["bundle_id"].(string); ok {
			bundleID = v
		}
		if v, ok := report.Metadata["platform"].(string); ok {
			platform = v
		}
		if v, ok := report.Metadata["version"].(string); ok {
			version = v
		}
		if v, ok := report.Metadata["git_branch"].(string); ok {
			branch = v
		}
		if v, ok := report.Metadata["git_commit"].(string); ok {
			commitSHA = v
		}
	}

	// Determine platform from artifact type if not set
	if platform == "" {
		switch report.ArtifactInfo.Type {
		case types.ArtifactTypeIPA, types.ArtifactTypeApp, types.ArtifactTypeXCArchive:
			platform = "iOS"
		case types.ArtifactTypeAPK, types.ArtifactTypeAAB:
			platform = "Android"
		}
	}

	return templateData{
		Title:              f.Title,
		AppName:            appName,
		BundleID:           bundleID,
		Platform:           platform,
		Version:            version,
		Branch:             branch,
		CommitSHA:          commitSHA,
		ArtifactName:       artifactName,
		ArtifactType:       string(report.ArtifactInfo.Type),
		TotalSize:          util.FormatBytes(downloadSize),
		UncompressedSize:   util.FormatBytes(uncompressedSize),
		CompressionRatio:   compressionRatio,
		TotalSavings:       util.FormatBytes(report.TotalSavings),
		SavingsPercentage:  savingsPercentage,
		Timestamp:          report.ArtifactInfo.AnalyzedAt.Format(time.RFC3339),
		DataJSON:           template.JS(dataJSON),
		NodeCount:          nodeCount,
		PerformanceWarning: performanceWarning,
	}
}

// jsData holds the data structure passed to JavaScript
type jsData struct {
	FileTree      interface{}           `json:"fileTree"`
	Categories    []categoryData        `json:"categories"`
	Extensions    []extensionData       `json:"extensions"`
	Optimizations []optimizationData    `json:"optimizations"`
	Duplicates    []string              `json:"duplicates"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type categoryData struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type extensionData struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type optimizationData struct {
	Category    string   `json:"category"`
	Severity    string   `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      int64    `json:"impact"`
	Files       []string `json:"files"`
	Action      string   `json:"action"`
}

// prepareJSData converts report data to JavaScript-friendly format
func (f *HTMLFormatter) prepareJSData(report *types.Report) jsData {
	return jsData{
		FileTree:      f.prepareTreemapData(report.FileTree),
		Categories:    f.prepareCategoryData(&report.SizeBreakdown),
		Extensions:    f.prepareExtensionData(&report.SizeBreakdown),
		Optimizations: f.prepareOptimizationData(report.Optimizations),
		Duplicates:    f.extractDuplicatePaths(report.Duplicates, report.ArtifactInfo.Path),
		Metadata:      report.Metadata,
	}
}

// extractDuplicatePaths collects all file paths that are duplicates
// It strips the artifact path prefix to match the file tree paths
func (f *HTMLFormatter) extractDuplicatePaths(duplicates []types.DuplicateSet, artifactPath string) []string {
	// Normalize artifact path for prefix stripping
	prefix := artifactPath
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	pathSet := make(map[string]struct{})
	for _, dup := range duplicates {
		for _, file := range dup.Files {
			// Strip artifact path prefix if present
			relativePath := file
			if strings.HasPrefix(file, prefix) {
				relativePath = file[len(prefix):]
			}
			pathSet[relativePath] = struct{}{}
		}
	}

	paths := make([]string, 0, len(pathSet))
	for path := range pathSet {
		paths = append(paths, path)
	}
	return paths
}

// prepareTreemapData converts FileNode tree to ECharts-compatible format
func (f *HTMLFormatter) prepareTreemapData(nodes []*types.FileNode) interface{} {
	if len(nodes) == 0 {
		return map[string]interface{}{
			"name":  "root",
			"value": 0,
		}
	}

	// If single root, use it; otherwise create a wrapper root
	if len(nodes) == 1 {
		return f.convertNodeToMap(nodes[0], 0)
	}

	// Multiple roots - create wrapper
	var totalSize int64
	children := make([]interface{}, 0, len(nodes))
	for _, node := range nodes {
		totalSize += node.Size
		children = append(children, f.convertNodeToMap(node, 0))
	}

	return map[string]interface{}{
		"name":     "root",
		"value":    totalSize,
		"path":     "/",
		"children": children,
	}
}

// convertNodeToMap recursively converts a FileNode to a map for ECharts
func (f *HTMLFormatter) convertNodeToMap(node *types.FileNode, depth int) map[string]interface{} {
	result := map[string]interface{}{
		"name":  node.Name,
		"value": node.Size,
		"path":  node.Path,
	}

	// Add file type for visual mapping
	if !node.IsDir {
		result["fileType"] = f.getFileType(node.Name)
	}

	// Process children
	if len(node.Children) > 0 {
		children := make([]interface{}, 0, len(node.Children))

		for _, child := range node.Children {
			children = append(children, f.convertNodeToMap(child, depth+1))
		}

		if len(children) > 0 {
			result["children"] = children
		}
	}

	return result
}

// getFileType determines the file type for visual mapping
func (f *HTMLFormatter) getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	lowerName := strings.ToLower(filename)

	// Check for framework
	if strings.HasSuffix(lowerName, ".framework") {
		return "framework"
	}

	typeMap := map[string]string{
		".dylib":      "library",
		".a":          "library",
		".so":         "native",
		".png":        "image",
		".jpg":        "image",
		".jpeg":       "image",
		".car":        "asset_catalog",
		".plist":      "resource",
		".storyboard": "ui",
		".xib":        "ui",
		".dex":        "dex",
		".xml":        "resource",
		".json":       "resource",
		".ttf":        "font",
		".otf":        "font",
	}

	if fileType, ok := typeMap[ext]; ok {
		return fileType
	}

	return "other"
}

// prepareCategoryData converts SizeBreakdown to category chart data
func (f *HTMLFormatter) prepareCategoryData(breakdown *types.SizeBreakdown) []categoryData {
	categories := []struct {
		name string
		size int64
	}{
		{"Frameworks", breakdown.Frameworks},
		{"Resources", breakdown.Resources},
		{"Executable", breakdown.Executable},
		{"Assets", breakdown.Assets},
		{"Libraries", breakdown.Libraries},
		{"DEX", breakdown.DEX},
		{"Other", breakdown.Other},
	}

	result := make([]categoryData, 0, len(categories))
	for _, cat := range categories {
		if cat.size > 0 {
			result = append(result, categoryData{
				Name:  cat.name,
				Value: cat.size,
			})
		}
	}

	// Sort by size descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})

	return result
}

// prepareExtensionData converts extension breakdown to chart data
func (f *HTMLFormatter) prepareExtensionData(breakdown *types.SizeBreakdown) []extensionData {
	if len(breakdown.ByExtension) == 0 {
		return []extensionData{}
	}

	// Convert to slice and sort
	result := make([]extensionData, 0, len(breakdown.ByExtension))
	for ext, size := range breakdown.ByExtension {
		result = append(result, extensionData{
			Name:  ext,
			Value: size,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})

	// Take top 10
	if len(result) > 10 {
		result = result[:10]
	}

	return result
}

// prepareOptimizationData converts optimizations to JavaScript format
func (f *HTMLFormatter) prepareOptimizationData(optimizations []types.Optimization) []optimizationData {
	result := make([]optimizationData, len(optimizations))
	for i, opt := range optimizations {
		result[i] = optimizationData{
			Category:    opt.Category,
			Severity:    opt.Severity,
			Title:       opt.Title,
			Description: opt.Description,
			Impact:      opt.Impact,
			Files:       opt.Files,
			Action:      opt.Action,
		}
	}
	return result
}

// calculateNodeCount counts total nodes in the tree
func (f *HTMLFormatter) calculateNodeCount(nodes []*types.FileNode) int {
	count := len(nodes)
	for _, node := range nodes {
		count += f.calculateNodeCount(node.Children)
	}
	return count
}

// executeTemplate writes the HTML output using the embedded template
func (f *HTMLFormatter) executeTemplate(w io.Writer, data templateData) error {
	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(w, data)
}
