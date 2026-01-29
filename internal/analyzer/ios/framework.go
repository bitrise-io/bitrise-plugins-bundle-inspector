package ios

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"howett.net/plist"
)

// FrameworkInfo contains metadata about a framework.
type FrameworkInfo struct {
	Name         string             `json:"name"`
	Path         string             `json:"path"`
	Version      string             `json:"version,omitempty"`
	Size         int64              `json:"size"`
	BinaryInfo   *macho.BinaryInfo  `json:"binary_info,omitempty"`
	Dependencies []string           `json:"dependencies,omitempty"`
}

// DiscoverFrameworks finds all .framework directories in the app bundle.
func DiscoverFrameworks(appPath string) ([]*FrameworkInfo, error) {
	var frameworks []*FrameworkInfo

	// Check for Frameworks directory
	frameworksDir := filepath.Join(appPath, "Frameworks")
	if _, err := os.Stat(frameworksDir); os.IsNotExist(err) {
		// No frameworks directory
		return frameworks, nil
	}

	// Walk the Frameworks directory
	err := filepath.Walk(frameworksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for .framework directories
		if info.IsDir() && strings.HasSuffix(path, ".framework") {
			fw, err := ParseFrameworkInfo(path, appPath)
			if err != nil {
				// Log warning but continue
				return nil
			}
			frameworks = append(frameworks, fw)
			// Don't recurse into framework
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk frameworks directory: %w", err)
	}

	return frameworks, nil
}

// ParseFrameworkInfo extracts metadata from a framework bundle.
func ParseFrameworkInfo(frameworkPath, appPath string) (*FrameworkInfo, error) {
	frameworkName := filepath.Base(frameworkPath)

	// Get framework size
	size, err := getDirectorySize(frameworkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get framework size: %w", err)
	}

	// Create relative path from app bundle
	relPath, err := filepath.Rel(appPath, frameworkPath)
	if err != nil {
		relPath = frameworkPath
	}

	info := &FrameworkInfo{
		Name: frameworkName,
		Path: relPath,
		Size: size,
	}

	// Try to get version from Info.plist
	infoPlistPath := filepath.Join(frameworkPath, "Info.plist")
	if version, err := GetFrameworkVersion(infoPlistPath); err == nil {
		info.Version = version
	}

	// Find and parse the framework binary
	// Framework binary is typically at <Framework>.framework/<Framework>
	binaryName := strings.TrimSuffix(frameworkName, ".framework")
	binaryPath := filepath.Join(frameworkPath, binaryName)

	if _, err := os.Stat(binaryPath); err == nil {
		// Try to parse as Mach-O
		if macho.IsMachO(binaryPath) {
			if binInfo, err := macho.ParseMachO(binaryPath); err == nil {
				info.BinaryInfo = binInfo
				// Store linked libraries as dependencies
				info.Dependencies = binInfo.LinkedLibraries
			}
		}
	}

	return info, nil
}

// GetFrameworkVersion reads CFBundleShortVersionString from Info.plist.
func GetFrameworkVersion(infoPlistPath string) (string, error) {
	data, err := os.ReadFile(infoPlistPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Info.plist: %w", err)
	}

	var plistData map[string]interface{}
	if _, err := plist.Unmarshal(data, &plistData); err != nil {
		return "", fmt.Errorf("failed to parse Info.plist: %w", err)
	}

	// Try CFBundleShortVersionString first
	if version, ok := plistData["CFBundleShortVersionString"].(string); ok {
		return version, nil
	}

	// Fallback to CFBundleVersion
	if version, ok := plistData["CFBundleVersion"].(string); ok {
		return version, nil
	}

	return "", fmt.Errorf("no version found in Info.plist")
}

// getDirectorySize calculates the total size of a directory.
func getDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}
