package detector

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/testutil"
)

func TestImageOptimizationDetector_Name(t *testing.T) {
	detector := NewImageOptimizationDetector(PlatformIOS)
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

	detector := NewImageOptimizationDetector(PlatformIOS)

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

	detector := NewImageOptimizationDetector(PlatformIOS)

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
	testutil.CreatePNGFile(t, tempDir, "small.png")  // < 5KB
	testutil.CreateJPEGFile(t, tempDir, "small.jpg") // < 10KB

	detector := NewImageOptimizationDetector(PlatformIOS)

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

	detector := NewImageOptimizationDetector(PlatformIOS)

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

// --- Android-specific tests ---

func TestImageOptimizationDetector_Android_NoToolRequired(t *testing.T) {
	// Android detector should work without sips or cwebp (uses estimation fallback)
	tempDir := testutil.CreateTempDir(t)

	// Create PNG files above the 5KB threshold
	testutil.CreateTestFile(t, tempDir, "icon.png", 6*1024)

	detector := NewImageOptimizationDetector(PlatformAndroid)

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Android detector should not require external tools, got error: %v", err)
	}

	// Should produce optimizations via estimation
	if len(optimizations) == 0 {
		t.Error("Expected at least one optimization from estimation fallback")
	}

	for _, opt := range optimizations {
		if opt.Category != "image-optimization" {
			t.Errorf("Expected category 'image-optimization', got %q", opt.Category)
		}
		if !strings.Contains(opt.Title, "WebP") {
			t.Errorf("Expected title to mention WebP, got %q", opt.Title)
		}
		if strings.Contains(opt.Title, "HEIC") {
			t.Errorf("Android optimization should not mention HEIC, got %q", opt.Title)
		}
		if !strings.Contains(opt.Description, "WebP") {
			t.Errorf("Expected description to mention WebP, got %q", opt.Description)
		}
		if !strings.Contains(opt.Action, "WebP") {
			t.Errorf("Expected action to mention WebP, got %q", opt.Action)
		}
	}
}

func TestImageOptimizationDetector_Android_SkipsWebP(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create a WebP file above threshold - should be skipped on Android
	testutil.CreateTestFile(t, tempDir, "already_optimized.webp", 11*1024)

	detector := NewImageOptimizationDetector(PlatformAndroid)

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for WebP files on Android, got %d", len(optimizations))
	}
}

func TestImageOptimizationDetector_Android_SmallFiles(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create files below thresholds
	testutil.CreatePNGFile(t, tempDir, "small.png")  // < 5KB
	testutil.CreateJPEGFile(t, tempDir, "small.jpg") // < 10KB

	detector := NewImageOptimizationDetector(PlatformAndroid)

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for small files, got %d", len(optimizations))
	}
}

func TestEstimateWebPSavings(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create a 10KB file
	filePath := testutil.CreateTestFile(t, tempDir, "test.png", 10*1024)

	savings, err := estimateWebPSavings(filePath)
	if err != nil {
		t.Fatalf("estimateWebPSavings failed: %v", err)
	}

	// Expected: 25% of 10KB = 2560 bytes
	expectedSavings := int64(float64(10*1024) * 0.25)
	if savings != expectedSavings {
		t.Errorf("Expected savings of %d bytes, got %d", expectedSavings, savings)
	}
}

func TestEstimateWebPSavings_NonexistentFile(t *testing.T) {
	_, err := estimateWebPSavings("/nonexistent/file.png")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestImageOptimizationDetector_TargetFormat(t *testing.T) {
	iosDetector := NewImageOptimizationDetector(PlatformIOS)
	if got := iosDetector.targetFormat(); got != "HEIC" {
		t.Errorf("iOS targetFormat: expected HEIC, got %q", got)
	}

	androidDetector := NewImageOptimizationDetector(PlatformAndroid)
	if got := androidDetector.targetFormat(); got != "WebP" {
		t.Errorf("Android targetFormat: expected WebP, got %q", got)
	}
}

func TestImageOptimizationDetector_ShouldOptimizeImage_Android(t *testing.T) {
	d := NewImageOptimizationDetector(PlatformAndroid)

	tests := []struct {
		ext            string
		wantOptimize   bool
		wantFormatName string
	}{
		{".png", true, "PNG"},
		{".jpg", true, "JPEG"},
		{".jpeg", true, "JPEG"},
		{".webp", false, ""},  // WebP already optimal on Android
		{".gif", false, ""},   // Not supported
		{".svg", false, ""},   // Not supported
	}

	for _, tt := range tests {
		shouldOptimize, formatName, _ := d.shouldOptimizeImage(tt.ext)
		if shouldOptimize != tt.wantOptimize {
			t.Errorf("shouldOptimizeImage(%q): got optimize=%v, want %v", tt.ext, shouldOptimize, tt.wantOptimize)
		}
		if formatName != tt.wantFormatName {
			t.Errorf("shouldOptimizeImage(%q): got formatName=%q, want %q", tt.ext, formatName, tt.wantFormatName)
		}
	}
}

func TestImageOptimizationDetector_ShouldOptimizeImage_iOS(t *testing.T) {
	d := NewImageOptimizationDetector(PlatformIOS)

	tests := []struct {
		ext            string
		wantOptimize   bool
		wantFormatName string
	}{
		{".png", true, "PNG"},
		{".jpg", true, "JPEG"},
		{".jpeg", true, "JPEG"},
		{".webp", true, "WebP"}, // WebP can be converted to HEIC on iOS
		{".gif", false, ""},
	}

	for _, tt := range tests {
		shouldOptimize, formatName, _ := d.shouldOptimizeImage(tt.ext)
		if shouldOptimize != tt.wantOptimize {
			t.Errorf("shouldOptimizeImage(%q): got optimize=%v, want %v", tt.ext, shouldOptimize, tt.wantOptimize)
		}
		if formatName != tt.wantFormatName {
			t.Errorf("shouldOptimizeImage(%q): got formatName=%q, want %q", tt.ext, formatName, tt.wantFormatName)
		}
	}
}
