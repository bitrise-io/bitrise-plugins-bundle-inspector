// Package bitrise provides Bitrise CI/CD integration utilities
package bitrise

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// BuildMetadata contains Bitrise build information
type BuildMetadata struct {
	BuildNumber string
	CommitHash  string
	DeployDir   string
}

// IsBitriseEnvironment checks if running in Bitrise CI
func IsBitriseEnvironment() bool {
	return os.Getenv("BITRISE_BUILD_NUMBER") != ""
}

// DetectBundlePath returns the bundle path from Bitrise environment variables
// Priority order: IPA > AAB > APK
func DetectBundlePath() (string, error) {
	// Check for iOS IPA
	if path := os.Getenv("BITRISE_IPA_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Check for Android AAB
	if path := os.Getenv("BITRISE_AAB_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Check for Android APK
	if path := os.Getenv("BITRISE_APK_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no bundle found in Bitrise environment variables (checked BITRISE_IPA_PATH, BITRISE_AAB_PATH, BITRISE_APK_PATH)")
}

// GetBuildMetadata returns Bitrise build information from environment variables
func GetBuildMetadata() BuildMetadata {
	return BuildMetadata{
		BuildNumber: os.Getenv("BITRISE_BUILD_NUMBER"),
		CommitHash:  os.Getenv("GIT_CLONE_COMMIT_HASH"),
		DeployDir:   os.Getenv("BITRISE_DEPLOY_DIR"),
	}
}

// ExportToDeployDir copies a file to the Bitrise deploy directory
// Returns the destination path or error
func ExportToDeployDir(sourcePath, filename string) (string, error) {
	metadata := GetBuildMetadata()
	if metadata.DeployDir == "" {
		return "", fmt.Errorf("BITRISE_DEPLOY_DIR not set")
	}

	// Ensure deploy directory exists
	if err := os.MkdirAll(metadata.DeployDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create deploy directory: %w", err)
	}

	destPath := filepath.Join(metadata.DeployDir, filename)

	// Open source file
	src, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy contents
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return destPath, nil
}

// WriteToDeployDir writes content to a file in the Bitrise deploy directory
// Returns the file path or error
func WriteToDeployDir(filename string, content []byte) (string, error) {
	metadata := GetBuildMetadata()
	if metadata.DeployDir == "" {
		return "", fmt.Errorf("BITRISE_DEPLOY_DIR not set")
	}

	// Ensure deploy directory exists
	if err := os.MkdirAll(metadata.DeployDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create deploy directory: %w", err)
	}

	destPath := filepath.Join(metadata.DeployDir, filename)

	// Write content to file
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return destPath, nil
}
