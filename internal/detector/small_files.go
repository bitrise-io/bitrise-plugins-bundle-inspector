package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// SmallFile represents a file smaller than the minimum block size
type SmallFile struct {
	Path       string
	Size       int64
	WastedSize int64 // 4KB - actual size
	Extension  string
}

// SmallFilesDetector implements the Detector interface
type SmallFilesDetector struct{}

// NewSmallFilesDetector creates a new small files detector
func NewSmallFilesDetector() *SmallFilesDetector {
	return &SmallFilesDetector{}
}

// Name returns the detector name
func (d *SmallFilesDetector) Name() string {
	return "small-files"
}

// shouldSkipSmallFileDetection determines if a file should be excluded from small file detection
func shouldSkipSmallFileDetection(path string) bool {
	// Skip files in frameworks (they're bundled as units)
	if strings.Contains(path, ".framework/") {
		return true
	}

	// Skip compiled asset catalogs (already optimized)
	if strings.Contains(path, "Assets.car") {
		return true
	}

	// Skip dylibs and binaries (can't be consolidated)
	if strings.HasSuffix(path, ".dylib") || strings.HasSuffix(path, ".so") {
		return true
	}

	// Skip compiled resources that can't be merged
	if strings.HasSuffix(path, ".nib") || strings.HasSuffix(path, ".storyboardc") {
		return true
	}

	return false
}

// DetectSmallFiles finds files smaller than the iOS block size (4KB)
func DetectSmallFiles(rootPath string) ([]SmallFile, error) {
	var smallFiles []SmallFile

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Skip files that shouldn't be included
		if shouldSkipSmallFileDetection(path) {
			return nil
		}

		// Check if file is smaller than block size
		if info.Size() < util.BlockSize && info.Size() > 0 {
			wastedSize := util.BlockSize - info.Size()
			smallFiles = append(smallFiles, SmallFile{
				Path:       path,
				Size:       info.Size(),
				WastedSize: wastedSize,
				Extension:  strings.ToLower(filepath.Ext(path)),
			})
		}

		return nil
	})

	return smallFiles, err
}

// groupByExtension groups small files by file extension
func groupByExtension(files []SmallFile) map[string][]SmallFile {
	groups := make(map[string][]SmallFile)
	for _, file := range files {
		ext := file.Extension
		if ext == "" {
			ext = "(no extension)"
		}
		groups[ext] = append(groups[ext], file)
	}
	return groups
}

// Detect runs the detector and returns optimizations for small files
func (d *SmallFilesDetector) Detect(rootPath string) ([]types.Optimization, error) {
	mapper := util.NewPathMapper(rootPath)

	// Detect all small files
	smallFiles, err := DetectSmallFiles(rootPath)
	if err != nil {
		return nil, WrapError("small-files", "detecting small files", err)
	}

	if len(smallFiles) == 0 {
		return nil, nil
	}

	// Group by extension
	groups := groupByExtension(smallFiles)

	var optimizations []types.Optimization

	// Create optimization for each extension group with significant waste
	for ext, files := range groups {
		// Only report if there are multiple files and significant waste
		if len(files) < 3 {
			continue // Skip small groups
		}

		var totalWasted int64
		for _, file := range files {
			totalWasted += file.WastedSize
		}

		// Only report if waste is significant (> 10KB)
		if totalWasted < 10*1024 {
			continue
		}

		// Sort files by size (smallest first) to show the most problematic files
		sort.Slice(files, func(i, j int) bool {
			return files[i].Size < files[j].Size
		})

		// Limit to top 20 smallest files to avoid overwhelming output
		displayCount := len(files)
		if displayCount > 20 {
			displayCount = 20
		}

		// Extract paths and calculate impact for displayed files only
		var displayFiles []string
		var displayedWaste int64
		for i := 0; i < displayCount; i++ {
			displayFiles = append(displayFiles, mapper.ToRelative(files[i].Path))
			displayedWaste += files[i].WastedSize
		}

		// Create descriptive title and action based on file type
		// Use totalWasted for title/description (shows total problem size)
		// But use displayedWaste for impact (matches displayed files)
		title, description, action := generateRecommendation(ext, len(files), totalWasted)

		optimizations = append(optimizations, types.Optimization{
			Category:    "small-files",
			Severity:    determineSeverity(len(files), totalWasted),
			Title:       title,
			Description: description,
			Impact:      displayedWaste, // Impact matches the displayed files
			Files:       displayFiles,
			Action:      action,
		})
	}

	// Sort by impact (highest first)
	sort.Slice(optimizations, func(i, j int) bool {
		return optimizations[i].Impact > optimizations[j].Impact
	})

	return optimizations, nil
}

// generateRecommendation creates tailored recommendations based on file type
func generateRecommendation(ext string, count int, totalWasted int64) (title, description, action string) {
	wastedStr := util.FormatBytes(totalWasted)

	switch ext {
	case ".strings", ".stringsdict":
		title = fmt.Sprintf("Consolidate %d localization files", count)
		description = fmt.Sprintf(
			"Found %d localization files smaller than 4KB, wasting %s due to iOS filesystem block size. "+
				"Each file smaller than 4KB still occupies a full 4KB block on disk.",
			count, wastedStr)
		action = "Merge small .strings files into fewer, larger localization files per language"

	case ".plist":
		title = fmt.Sprintf("Consolidate %d property list files", count)
		description = fmt.Sprintf(
			"Found %d .plist files smaller than 4KB, wasting %s. "+
				"Small property lists can often be merged or embedded in code as constants.",
			count, wastedStr)
		action = "Merge related .plist files or convert small configurations to in-code constants"

	case ".json":
		title = fmt.Sprintf("Consolidate %d JSON configuration files", count)
		description = fmt.Sprintf(
			"Found %d JSON files smaller than 4KB, wasting %s. "+
				"Consider bundling related JSON files into a single configuration file.",
			count, wastedStr)
		action = "Merge related JSON files into a single configuration bundle"

	case ".png", ".jpg", ".jpeg", ".gif":
		title = fmt.Sprintf("Move %d small images to asset catalog", count)
		description = fmt.Sprintf(
			"Found %d small image files (<4KB) wasting %s. "+
				"Asset catalogs can pack small images more efficiently.",
			count, wastedStr)
		action = "Move small images into .xcassets asset catalog for better compression and efficiency"

	default:
		title = fmt.Sprintf("Consolidate %d small %s files", count, ext)
		description = fmt.Sprintf(
			"Found %d %s files smaller than 4KB, wasting %s. "+
				"iOS allocates a minimum of 4KB per file regardless of content size.",
			count, ext, wastedStr)
		action = "Consolidate small files of the same type into fewer, larger files"
	}

	return title, description, action
}

// determineSeverity returns severity based on file count and wasted space
func determineSeverity(count int, totalWasted int64) string {
	// High severity: many files or significant waste
	if count >= 100 || totalWasted >= 100*1024 {
		return "high"
	}
	// Medium severity: moderate files or waste
	if count >= 20 || totalWasted >= 20*1024 {
		return "medium"
	}
	// Low severity: few files or minimal waste
	return "low"
}
