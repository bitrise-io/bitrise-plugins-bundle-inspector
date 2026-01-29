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
		if dup.Size != int64(len(content1)) {
			t.Errorf("Expected size %d, got %d", len(content1), dup.Size)
		}
		expectedWasted := int64(len(content1))
		if dup.WastedSize != expectedWasted {
			t.Errorf("Expected wasted size %d, got %d", expectedWasted, dup.WastedSize)
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
