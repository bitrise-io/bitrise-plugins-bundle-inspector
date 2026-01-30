package detector

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/testutil"
)

func TestDetectUnnecessaryFiles(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create test files
	testutil.CreateTestFile(t, tempDir, "Framework.framework/module.modulemap", 100)
	testutil.CreateTestFile(t, tempDir, "Framework.framework/Headers/Test.h", 200)
	testutil.CreateTestFile(t, tempDir, "Framework.framework/Headers/Test.hpp", 150)
	testutil.CreateTestFile(t, tempDir, "Framework.framework/Modules/Test.swiftmodule", 300)
	testutil.CreateTestFile(t, tempDir, "Framework.framework/Modules/Test.swiftdoc", 250)
	testutil.CreateTestFile(t, tempDir, "README.md", 500)
	testutil.CreateTestFile(t, tempDir, "CHANGELOG.md", 400)
	testutil.CreateTestFile(t, tempDir, ".gitkeep", 0)
	testutil.CreateTestFile(t, tempDir, "ValidFile.swift", 1000) // Should not be detected

	unnecessary, err := DetectUnnecessaryFiles(tempDir)
	if err != nil {
		t.Fatalf("DetectUnnecessaryFiles failed: %v", err)
	}

	if len(unnecessary) != 8 {
		t.Errorf("Expected 8 unnecessary files, got %d", len(unnecessary))
		for _, file := range unnecessary {
			t.Logf("Found: %s", file.Path)
		}
	}

	// Verify specific files were detected
	expectedFiles := map[string]bool{
		"module.modulemap": false,
		"Test.h":           false,
		"Test.hpp":         false,
		"Test.swiftmodule": false,
		"Test.swiftdoc":    false,
		"README.md":        false,
		"CHANGELOG.md":     false,
		".gitkeep":         false,
	}

	for _, file := range unnecessary {
		for expectedFile := range expectedFiles {
			if contains(file.Path, expectedFile) {
				expectedFiles[expectedFile] = true
			}
		}
	}

	for file, found := range expectedFiles {
		if !found {
			t.Errorf("Expected file %s not found in unnecessary files", file)
		}
	}
}

func TestGetRemovalReason(t *testing.T) {
	tests := []struct {
		pattern string
		want    string
	}{
		{
			pattern: "module.modulemap",
			want:    "Clang module map not needed in release builds",
		},
		{
			pattern: ".swiftmodule",
			want:    "Swift module/doc files not needed in release builds",
		},
		{
			pattern: ".swiftdoc",
			want:    "Swift module/doc files not needed in release builds",
		},
		{
			pattern: ".h",
			want:    "Header files not needed in release builds",
		},
		{
			pattern: ".hpp",
			want:    "Header files not needed in release builds",
		},
		{
			pattern: "README.md",
			want:    "Documentation files not needed in release builds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := getRemovalReason(tt.pattern)
			if got != tt.want {
				t.Errorf("getRemovalReason(%q) = %q, want %q", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestUnnecessaryFilesDetector_Detect(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create test files
	testutil.CreateTestFile(t, tempDir, "module.modulemap", 100)
	testutil.CreateTestFile(t, tempDir, "Test.h", 200)
	testutil.CreateTestFile(t, tempDir, "README.md", 500)

	detector := NewUnnecessaryFilesDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(optimizations) == 0 {
		t.Error("Expected optimizations, got none")
	}

	// Verify optimization structure
	for _, opt := range optimizations {
		if opt.Category != "unnecessary-files" {
			t.Errorf("Expected category 'unnecessary-files', got %q", opt.Category)
		}
		if opt.Severity != "low" {
			t.Errorf("Expected severity 'low', got %q", opt.Severity)
		}
		if opt.Impact == 0 {
			t.Error("Expected non-zero impact")
		}
		if len(opt.Files) == 0 {
			t.Error("Expected files list to be non-empty")
		}
	}
}

func TestUnnecessaryFilesDetector_Name(t *testing.T) {
	detector := NewUnnecessaryFilesDetector()
	if detector.Name() != "unnecessary-files" {
		t.Errorf("Expected name 'unnecessary-files', got %q", detector.Name())
	}
}

func TestDetectUnnecessaryFiles_EmptyDirectory(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	unnecessary, err := DetectUnnecessaryFiles(tempDir)
	if err != nil {
		t.Fatalf("DetectUnnecessaryFiles failed: %v", err)
	}

	if len(unnecessary) != 0 {
		t.Errorf("Expected no unnecessary files in empty directory, got %d", len(unnecessary))
	}
}

func TestUnnecessaryFilesDetector_Detect_EmptyResult(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create only valid files
	testutil.CreateTestFile(t, tempDir, "ValidFile.swift", 1000)
	testutil.CreateTestFile(t, tempDir, "AnotherFile.m", 500)

	detector := NewUnnecessaryFilesDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for valid files, got %d", len(optimizations))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[len(s)-len(substr):] == substr || findInPath(s, substr))
}

func findInPath(path, target string) bool {
	for i := 0; i <= len(path)-len(target); i++ {
		if path[i:i+len(target)] == target {
			return true
		}
	}
	return false
}
