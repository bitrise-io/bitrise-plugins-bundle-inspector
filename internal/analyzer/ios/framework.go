package ios

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
	"howett.net/plist"
)

// FrameworkInfo contains metadata about a framework.
type FrameworkInfo struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Version      string            `json:"version,omitempty"`
	Size         int64             `json:"size"`
	BinaryInfo   *types.BinaryInfo `json:"binary_info,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
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

// AppMetadata contains key metadata extracted from Info.plist
type AppMetadata struct {
	AppName      string
	BundleID     string
	Version      string
	BuildVersion string
	MinOSVersion string
	IconNames    []string // Icon base names from CFBundleIcons/CFBundleIconFiles
}

// ParseAppInfoPlist extracts app metadata from the main app's Info.plist
func ParseAppInfoPlist(infoPlistPath string) (*AppMetadata, error) {
	data, err := os.ReadFile(infoPlistPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Info.plist: %w", err)
	}

	var plistData map[string]interface{}
	if _, err := plist.Unmarshal(data, &plistData); err != nil {
		return nil, fmt.Errorf("failed to parse Info.plist: %w", err)
	}

	metadata := &AppMetadata{}

	// Extract app name (try multiple keys)
	if name, ok := plistData["CFBundleDisplayName"].(string); ok && name != "" {
		metadata.AppName = name
	} else if name, ok := plistData["CFBundleName"].(string); ok && name != "" {
		metadata.AppName = name
	}

	// Extract bundle identifier
	if bundleID, ok := plistData["CFBundleIdentifier"].(string); ok {
		metadata.BundleID = bundleID
	}

	// Extract version
	if version, ok := plistData["CFBundleShortVersionString"].(string); ok {
		metadata.Version = version
	}

	// Extract build version
	if buildVersion, ok := plistData["CFBundleVersion"].(string); ok {
		metadata.BuildVersion = buildVersion
	}

	// Extract minimum OS version
	if minOS, ok := plistData["MinimumOSVersion"].(string); ok {
		metadata.MinOSVersion = minOS
	}

	// Extract icon names
	metadata.IconNames = extractIconNames(plistData)

	return metadata, nil
}

// extractIconNames extracts icon base names from a parsed Info.plist.
// It checks CFBundleIcons, CFBundleIcons~ipad, and legacy keys.
// Returns nil if no icon names are declared (common for apps using Assets.car only).
func extractIconNames(plistData map[string]interface{}) []string {
	var names []string

	// CFBundleIcons > CFBundlePrimaryIcon > CFBundleIconFiles / CFBundleIconName
	if icons, ok := plistData["CFBundleIcons"].(map[string]interface{}); ok {
		names = append(names, extractIconFilesFromDict(icons)...)
	}

	// CFBundleIcons~ipad variant
	if icons, ok := plistData["CFBundleIcons~ipad"].(map[string]interface{}); ok {
		names = append(names, extractIconFilesFromDict(icons)...)
	}

	// Legacy: top-level CFBundleIconFiles (array)
	if iconFiles, ok := plistData["CFBundleIconFiles"].([]interface{}); ok {
		for _, f := range iconFiles {
			if s, ok := f.(string); ok && s != "" {
				names = append(names, s)
			}
		}
	}

	// Legacy: CFBundleIconFile (singular, very old apps)
	if iconFile, ok := plistData["CFBundleIconFile"].(string); ok && iconFile != "" {
		names = append(names, iconFile)
	}

	// Normalize: strip .png extension (legacy plists may include it, e.g. "Icon.png")
	for i, name := range names {
		names[i] = strings.TrimSuffix(name, ".png")
	}

	return deduplicateStrings(names)
}

// extractIconFilesFromDict extracts icon names from a CFBundleIcons dictionary.
func extractIconFilesFromDict(icons map[string]interface{}) []string {
	var names []string

	primary, ok := icons["CFBundlePrimaryIcon"].(map[string]interface{})
	if !ok {
		return nil
	}

	// CFBundleIconFiles array (e.g., ["AppIcon60x60", "AppIcon76x76"])
	if iconFiles, ok := primary["CFBundleIconFiles"].([]interface{}); ok {
		for _, f := range iconFiles {
			if s, ok := f.(string); ok && s != "" {
				names = append(names, s)
			}
		}
	}

	// CFBundleIconName string (e.g., "AppIcon")
	if iconName, ok := primary["CFBundleIconName"].(string); ok && iconName != "" {
		names = append(names, iconName)
	}

	return names
}

// deduplicateStrings returns a new slice with duplicates removed, preserving order.
func deduplicateStrings(input []string) []string {
	if len(input) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))
	for _, s := range input {
		if _, exists := seen[s]; !exists {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
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
