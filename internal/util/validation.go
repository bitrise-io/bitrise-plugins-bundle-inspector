package util

import (
	"fmt"
	"os"
	"strings"
)

// validateArtifact is a common helper for validating artifact paths.
func validateArtifact(path, expectedExt string, shouldBeDir bool) error {
	// Normalize extension (ensure it starts with a dot)
	if !strings.HasPrefix(expectedExt, ".") {
		expectedExt = "." + expectedExt
	}

	// Check extension
	if !strings.HasSuffix(strings.ToLower(path), strings.ToLower(expectedExt)) {
		artifactType := "file"
		if shouldBeDir {
			artifactType = "path"
		}
		return fmt.Errorf("%s must have %s extension", artifactType, expectedExt)
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		artifactType := "file"
		if shouldBeDir {
			artifactType = "path"
		}
		return fmt.Errorf("failed to stat %s: %w", artifactType, err)
	}

	// Check if it matches expected type
	if shouldBeDir {
		if !info.IsDir() {
			return fmt.Errorf("path is a file, expected a %s directory", expectedExt)
		}
	} else {
		if info.IsDir() {
			return fmt.Errorf("path is a directory, expected a %s file", expectedExt)
		}
	}

	return nil
}

// ValidateFileArtifact validates that a path points to a file with the expected extension.
// It checks that:
// 1. The path has the expected extension (case-insensitive)
// 2. The path exists and is accessible
// 3. The path is a file, not a directory
func ValidateFileArtifact(path, expectedExt string) error {
	return validateArtifact(path, expectedExt, false)
}

// ValidateDirectoryArtifact validates that a path points to a directory with the expected extension.
// It checks that:
// 1. The path has the expected extension (e.g., .app)
// 2. The path exists and is accessible
// 3. The path is a directory, not a file
func ValidateDirectoryArtifact(path, expectedExt string) error {
	return validateArtifact(path, expectedExt, true)
}
