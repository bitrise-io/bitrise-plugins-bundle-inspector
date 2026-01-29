package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

var unnecessaryPatterns = []string{
	"module.modulemap", // Clang module maps (not needed in release)
	".swiftmodule",     // Swift module files (not needed in release)
	".swiftdoc",        // Swift documentation (not needed)
	".h",               // Header files (not needed in release)
	".hpp",             // C++ header files
	"README.md",        // Documentation
	"CHANGELOG.md",     // Documentation
	".gitkeep",         // Git placeholder
}

// UnnecessaryFile represents a file that shouldn't be in production bundle
type UnnecessaryFile struct {
	Path   string
	Size   int64
	Reason string
}

// DetectUnnecessaryFiles finds files that shouldn't be in production bundle
func DetectUnnecessaryFiles(rootPath string) ([]UnnecessaryFile, error) {
	var unnecessary []UnnecessaryFile

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		filename := filepath.Base(path)
		ext := strings.ToLower(filepath.Ext(path))

		for _, pattern := range unnecessaryPatterns {
			if pattern == filename || pattern == ext {
				reason := getRemovalReason(pattern)
				unnecessary = append(unnecessary, UnnecessaryFile{
					Path:   path,
					Size:   info.Size(),
					Reason: reason,
				})
				break
			}
		}

		return nil
	})

	return unnecessary, err
}

func getRemovalReason(pattern string) string {
	switch pattern {
	case "module.modulemap":
		return "Clang module map not needed in release builds"
	case ".swiftmodule", ".swiftdoc":
		return "Swift module/doc files not needed in release builds"
	case ".h", ".hpp":
		return "Header files not needed in release builds"
	default:
		return "Documentation files not needed in release builds"
	}
}

// UnnecessaryFilesDetector implements the Detector interface
type UnnecessaryFilesDetector struct{}

// NewUnnecessaryFilesDetector creates a new unnecessary files detector
func NewUnnecessaryFilesDetector() *UnnecessaryFilesDetector {
	return &UnnecessaryFilesDetector{}
}

// Name returns the detector name
func (d *UnnecessaryFilesDetector) Name() string {
	return "unnecessary-files"
}

// Detect runs the detector and returns optimizations grouped by type
func (d *UnnecessaryFilesDetector) Detect(rootPath string) ([]types.Optimization, error) {
	mapper := NewPathMapper(rootPath)
	unnecessary, err := DetectUnnecessaryFiles(rootPath)
	if err != nil || len(unnecessary) == 0 {
		return nil, err
	}

	// Group by reason
	grouped := make(map[string]struct {
		files []string
		size  int64
	})

	for _, file := range unnecessary {
		entry := grouped[file.Reason]
		entry.files = append(entry.files, mapper.ToRelative(file.Path))
		entry.size += file.Size
		grouped[file.Reason] = entry
	}

	// Create optimizations
	var optimizations []types.Optimization
	for reason, data := range grouped {
		if data.size < 1024 {
			continue // Skip if too small
		}

		optimizations = append(optimizations, types.Optimization{
			Category:    "unnecessary-files",
			Severity:    "low",
			Title:       fmt.Sprintf("Remove %d unnecessary files", len(data.files)),
			Description: reason,
			Impact:      data.size,
			Files:       data.files,
			Action:      "Remove these files from the bundle",
		})
	}

	return optimizations, nil
}
