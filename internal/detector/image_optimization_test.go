package detector

import (
	"os/exec"
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/testutil"
)

func TestImageOptimizationDetector_Name(t *testing.T) {
	detector := NewImageOptimizationDetector()
	if detector.Name() != "image-optimization" {
		t.Errorf("Expected name 'image-optimization', got %q", detector.Name())
	}
}

func TestCheckSipsAvailable(t *testing.T) {
	err := checkSipsAvailable()

	// Check if sips is actually available on this system
	_, lookPathErr := exec.LookPath("sips")

	if lookPathErr == nil {
		// sips should be available
		if err != nil {
			t.Errorf("checkSipsAvailable() returned error on system with sips: %v", err)
		}
	} else {
		// sips should not be available
		if err == nil {
			t.Error("checkSipsAvailable() returned nil on system without sips")
		}
	}
}

func TestImageOptimizationDetector_Detect_NoSips(t *testing.T) {
	// Skip this test if sips is available (we're testing the no-sips path)
	if _, err := exec.LookPath("sips"); err == nil {
		t.Skip("Skipping no-sips test on system with sips available")
	}

	tempDir := testutil.CreateTempDir(t)
	testutil.CreatePNGFile(t, tempDir, "large.png")

	detector := NewImageOptimizationDetector()

	_, err := detector.Detect(tempDir)
	if err == nil {
		t.Error("Expected error when sips is not available, got nil")
	}
}

func TestImageOptimizationDetector_Detect_WithSips(t *testing.T) {
	// Skip this test if sips is not available
	if _, err := exec.LookPath("sips"); err != nil {
		t.Skip("Skipping sips-dependent test: sips not available on this system")
	}

	tempDir := testutil.CreateTempDir(t)

	// Create test images with sufficient size to trigger optimization
	// PNG needs to be > 5KB
	testutil.CreateTestFile(t, tempDir, "large.png", 6*1024)
	// JPEG needs to be > 10KB
	testutil.CreateTestFile(t, tempDir, "photo.jpg", 11*1024)

	detector := NewImageOptimizationDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Note: Actual conversions might fail because test files aren't valid images
	// We're mainly testing that the detector runs without crashing
	t.Logf("Found %d optimizations (may be 0 if test files aren't valid images)", len(optimizations))

	// Verify optimization structure if any were found
	for _, opt := range optimizations {
		if opt.Category != "image-optimization" {
			t.Errorf("Expected category 'image-optimization', got %q", opt.Category)
		}
		if opt.Severity != "medium" {
			t.Errorf("Expected severity 'medium', got %q", opt.Severity)
		}
		if len(opt.Files) == 0 {
			t.Error("Expected files list to be non-empty")
		}
	}
}

func TestImageOptimizationDetector_Detect_SmallFiles(t *testing.T) {
	// Skip this test if sips is not available
	if _, err := exec.LookPath("sips"); err != nil {
		t.Skip("Skipping sips-dependent test: sips not available on this system")
	}

	tempDir := testutil.CreateTempDir(t)

	// Create small files that should be ignored
	testutil.CreatePNGFile(t, tempDir, "small.png")     // < 5KB
	testutil.CreateJPEGFile(t, tempDir, "small.jpg")    // < 10KB

	detector := NewImageOptimizationDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Small files should not generate optimizations
	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for small files, got %d", len(optimizations))
	}
}

func TestImageOptimizationDetector_Detect_EmptyDirectory(t *testing.T) {
	// Skip this test if sips is not available
	if _, err := exec.LookPath("sips"); err != nil {
		t.Skip("Skipping sips-dependent test: sips not available on this system")
	}

	tempDir := testutil.CreateTempDir(t)

	detector := NewImageOptimizationDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for empty directory, got %d", len(optimizations))
	}
}

func TestHasAlpha_NoSips(t *testing.T) {
	// Skip this test if sips is available
	if _, err := exec.LookPath("sips"); err == nil {
		t.Skip("Skipping no-sips test on system with sips available")
	}

	result := hasAlpha("/fake/path.png")
	if result != false {
		t.Error("Expected hasAlpha to return false when sips is not available")
	}
}

func TestMeasureActualHEICConversion_InvalidFile(t *testing.T) {
	// Skip this test if sips is not available
	if _, err := exec.LookPath("sips"); err != nil {
		t.Skip("Skipping sips-dependent test: sips not available on this system")
	}

	tempDir := testutil.CreateTempDir(t)
	invalidFile := testutil.CreateTestFile(t, tempDir, "invalid.png", 1000)

	_, err := measureActualHEICConversion(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid image file, got nil")
	}
}

func TestMeasureActualHEICConversion_NonexistentFile(t *testing.T) {
	// Skip this test if sips is not available
	if _, err := exec.LookPath("sips"); err != nil {
		t.Skip("Skipping sips-dependent test: sips not available on this system")
	}

	_, err := measureActualHEICConversion("/nonexistent/file.png")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestDetectImageOptimizations(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create test files
	testutil.CreateTestFile(t, tempDir, "large.png", 6*1024)  // > 5KB
	testutil.CreateTestFile(t, tempDir, "photo.jpg", 60*1024) // > 50KB
	testutil.CreateTestFile(t, tempDir, "small.png", 1*1024)  // < 5KB (should be ignored)

	optimizations, err := DetectImageOptimizations(tempDir)
	if err != nil {
		t.Fatalf("DetectImageOptimizations failed: %v", err)
	}

	// Should detect 2 optimization opportunities (large.png and photo.jpg)
	if len(optimizations) != 2 {
		t.Errorf("Expected 2 optimization opportunities, got %d", len(optimizations))
	}

	// Verify structure
	for _, opt := range optimizations {
		if opt.CurrentSize == 0 {
			t.Error("Expected non-zero current size")
		}
		if opt.EstimatedSavings == 0 {
			t.Error("Expected non-zero estimated savings")
		}
		if opt.Path == "" {
			t.Error("Expected non-empty path")
		}
		if opt.CurrentFormat == "" {
			t.Error("Expected non-empty current format")
		}
		if opt.RecommendedFormat == "" {
			t.Error("Expected non-empty recommended format")
		}
	}
}

func TestDetectImageOptimizations_EmptyDirectory(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	optimizations, err := DetectImageOptimizations(tempDir)
	if err != nil {
		t.Fatalf("DetectImageOptimizations failed: %v", err)
	}

	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for empty directory, got %d", len(optimizations))
	}
}
