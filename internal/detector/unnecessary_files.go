package detector

import (
	"os"
	"path/filepath"
	"strings"
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
