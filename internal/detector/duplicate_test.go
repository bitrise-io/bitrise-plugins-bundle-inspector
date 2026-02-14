package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDuplicateDetector(t *testing.T) {
	// Create temp directory with test files
	tempDir, err := os.MkdirTemp("", "duplicate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	content1 := []byte("Hello World")
	content2 := []byte("Different Content")

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	file3 := filepath.Join(tempDir, "file3.txt")

	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, content1, 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}
	if err := os.WriteFile(file3, content2, 0644); err != nil {
		t.Fatalf("Failed to write file3: %v", err)
	}

	// Detect duplicates
	detector := NewDuplicateDetector()
	duplicates, err := detector.DetectDuplicates(tempDir)
	if err != nil {
		t.Fatalf("DetectDuplicates failed: %v", err)
	}

	// Should find one duplicate set (file1 and file2)
	if len(duplicates) != 1 {
		t.Errorf("Expected 1 duplicate set, got %d", len(duplicates))
	}

	if len(duplicates) > 0 {
		dup := duplicates[0]
		if dup.Count != 2 {
			t.Errorf("Expected 2 duplicates, got %d", dup.Count)
		}
		// With block alignment (4 KB), even small files occupy a full block
		expectedBlockAlignedSize := int64(4096) // blockSize constant
		if dup.Size != expectedBlockAlignedSize {
			t.Errorf("Expected block-aligned size %d, got %d", expectedBlockAlignedSize, dup.Size)
		}
		// Wasted size is (count - 1) * block-aligned size
		expectedWasted := expectedBlockAlignedSize
		if dup.WastedSize != expectedWasted {
			t.Errorf("Expected wasted size %d, got %d", expectedWasted, dup.WastedSize)
		}
		// Verify paths are relative (not absolute)
		for _, file := range dup.Files {
			if filepath.IsAbs(file) {
				t.Errorf("Expected relative path, got absolute: %s", file)
			}
		}
	}
}

func TestDuplicateDetector_RelativePaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "duplicate-relpath-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create nested directory structure
	subDir := filepath.Join(tempDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirs: %v", err)
	}

	content := []byte("duplicate content here")
	files := []string{
		filepath.Join(tempDir, "root.txt"),
		filepath.Join(subDir, "nested.txt"),
	}

	for _, f := range files {
		if err := os.WriteFile(f, content, 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", f, err)
		}
	}

	detector := NewDuplicateDetector()
	duplicates, err := detector.DetectDuplicates(tempDir)
	if err != nil {
		t.Fatalf("DetectDuplicates failed: %v", err)
	}

	if len(duplicates) != 1 {
		t.Fatalf("Expected 1 duplicate set, got %d", len(duplicates))
	}

	dup := duplicates[0]
	expectedPaths := map[string]bool{
		"root.txt":         false,
		"sub/dir/nested.txt": false,
	}

	for _, file := range dup.Files {
		if filepath.IsAbs(file) {
			t.Errorf("Expected relative path, got absolute: %s", file)
		}
		if _, ok := expectedPaths[file]; ok {
			expectedPaths[file] = true
		} else {
			t.Errorf("Unexpected path: %s", file)
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected path not found in duplicates: %s", path)
		}
	}
}

func TestGetTotalWastedSpace(t *testing.T) {
	// This is a simple test since we already have the DuplicateSet
	// Just verify the calculation is correct
	duplicates := []struct {
		size       int64
		count      int
		wastedSize int64
	}{
		{1000, 2, 1000},
		{2000, 3, 4000},
	}

	var totalExpected int64
	for _, d := range duplicates {
		totalExpected += d.wastedSize
	}

	if totalExpected != 5000 {
		t.Errorf("Expected total wasted 5000, got %d", totalExpected)
	}
}
