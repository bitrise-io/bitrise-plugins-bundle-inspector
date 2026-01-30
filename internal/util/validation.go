package util

import (
	"fmt"
	"os"
	"strings"
)

// ValidateFileArtifact validates that a path points to a file with the expected extension.
// It checks that:
// 1. The path has the expected extension (case-insensitive)
// 2. The path exists and is accessible
// 3. The path is a file, not a directory
func ValidateFileArtifact(path, expectedExt string) error {
	// Normalize extension (ensure it starts with a dot)
	if !strings.HasPrefix(expectedExt, ".") {
		expectedExt = "." + expectedExt
	}

	// Check extension
	if !strings.HasSuffix(strings.ToLower(path), strings.ToLower(expectedExt)) {
		return fmt.Errorf("file must have %s extension", expectedExt)
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if it's a file (not a directory)
	if info.IsDir() {
		return fmt.Errorf("path is a directory, expected a %s file", expectedExt)
	}

	return nil
}

// ValidateDirectoryArtifact validates that a path points to a directory with the expected extension.
// It checks that:
// 1. The path has the expected extension (e.g., .app)
// 2. The path exists and is accessible
// 3. The path is a directory, not a file
func ValidateDirectoryArtifact(path, expectedExt string) error {
	// Normalize extension (ensure it starts with a dot)
	if !strings.HasPrefix(expectedExt, ".") {
		expectedExt = "." + expectedExt
	}

	// Check extension
	if !strings.HasSuffix(strings.ToLower(path), strings.ToLower(expectedExt)) {
		return fmt.Errorf("path must have %s extension", expectedExt)
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	// Check if it's a directory (not a file)
	if !info.IsDir() {
		return fmt.Errorf("path is a file, expected a %s directory", expectedExt)
	}

	return nil
}
