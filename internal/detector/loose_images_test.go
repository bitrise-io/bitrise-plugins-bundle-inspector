package detector

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/testutil"
)

func TestDetectLooseImages(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create loose images (not in asset catalogs)
	testutil.CreatePNGFile(t, tempDir, "icon.png")
	testutil.CreateJPEGFile(t, tempDir, "background.jpg")
	testutil.CreatePNGFile(t, tempDir, "logo.png")

	// Create asset catalog
	catalogDir := testutil.CreateTestDir(t, tempDir, "Assets.xcassets")
	testutil.CreatePNGFile(t, catalogDir, "AppIcon.png")

	// Create Assets.car (compiled asset catalog)
	testutil.CreateTestFile(t, tempDir, "Assets.car", 1000)

	looseImages, err := DetectLooseImages(tempDir)
	if err != nil {
		t.Fatalf("DetectLooseImages failed: %v", err)
	}

	// Should find 3 loose images (not the ones in xcassets or Assets.car)
	if len(looseImages) != 3 {
		t.Errorf("Expected 3 loose images, got %d", len(looseImages))
		for _, img := range looseImages {
			t.Logf("Found: %s", img.Path)
		}
	}

	// Verify none are marked as in asset catalog
	for _, img := range looseImages {
		if img.InAssetCatalog {
			t.Errorf("Image %s should not be marked as in asset catalog", img.Path)
		}
	}
}

func TestExtractBaseName(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{
			filename: "icon@2x.png",
			want:     "icon",
		},
		{
			filename: "icon@3x.png",
			want:     "icon",
		},
		{
			filename: "icon@1x.png",
			want:     "icon",
		},
		{
			filename: "icon.png",
			want:     "icon",
		},
		{
			filename: "background@2x.jpg",
			want:     "background",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := extractBaseName(tt.filename)
			if got != tt.want {
				t.Errorf("extractBaseName(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestDetectPatternType(t *testing.T) {
	tests := []struct {
		name   string
		images []LooseImage
		want   string
	}{
		{
			name: "retina variants",
			images: []LooseImage{
				{Path: "/tmp/icon@1x.png", Size: 100},
				{Path: "/tmp/icon@2x.png", Size: 400},
				{Path: "/tmp/icon@3x.png", Size: 900},
			},
			want: "retina-variants",
		},
		{
			name: "multi-location with different sizes",
			images: []LooseImage{
				{Path: "/tmp/dir1/icon.png", Size: 100},
				{Path: "/tmp/dir2/icon.png", Size: 200},
			},
			want: "multi-location",
		},
		{
			name: "no pattern - same size",
			images: []LooseImage{
				{Path: "/tmp/dir1/icon.png", Size: 100},
				{Path: "/tmp/dir2/icon.png", Size: 100},
			},
			want: "",
		},
		{
			name: "no pattern - single image",
			images: []LooseImage{
				{Path: "/tmp/icon.png", Size: 100},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPatternType(tt.images)
			if got != tt.want {
				t.Errorf("detectPatternType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectPatterns(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create retina variant images
	testutil.CreateTestFile(t, tempDir, "icon@1x.png", 100)
	testutil.CreateTestFile(t, tempDir, "icon@2x.png", 400)
	testutil.CreateTestFile(t, tempDir, "icon@3x.png", 900)

	looseImages := []LooseImage{
		{Path: tempDir + "/icon@1x.png", Size: 100},
		{Path: tempDir + "/icon@2x.png", Size: 400},
		{Path: tempDir + "/icon@3x.png", Size: 900},
	}

	patterns := detectPatterns(looseImages)

	if len(patterns) != 1 {
		t.Errorf("Expected 1 pattern, got %d", len(patterns))
	}

	if len(patterns) > 0 {
		pattern := patterns[0]
		if pattern.patternType != "retina-variants" {
			t.Errorf("Expected pattern type 'retina-variants', got %q", pattern.patternType)
		}
		if len(pattern.images) != 3 {
			t.Errorf("Expected 3 images in pattern, got %d", len(pattern.images))
		}
	}
}

func TestCalculatePatternSavings(t *testing.T) {
	tests := []struct {
		name    string
		pattern imagePattern
		want    int64
	}{
		{
			name: "retina variants - save @1x and @2x",
			pattern: imagePattern{
				patternType: "retina-variants",
				images: []LooseImage{
					{Size: 1000},  // @1x
					{Size: 4000},  // @2x
					{Size: 9000},  // @3x (keep this)
				},
			},
			want: 4096 + 4096, // Both round up to 4KB blocks
		},
		{
			name: "multi-location - save smaller files",
			pattern: imagePattern{
				patternType: "multi-location",
				images: []LooseImage{
					{Size: 1000},
					{Size: 5000}, // Keep largest
				},
			},
			want: 4096, // 1000 bytes rounds up to 4KB
		},
		{
			name: "unknown pattern type",
			pattern: imagePattern{
				patternType: "unknown",
				images: []LooseImage{
					{Size: 1000},
					{Size: 2000},
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePatternSavings(tt.pattern)
			if got != tt.want {
				t.Errorf("calculatePatternSavings() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLooseImagesDetector_Name(t *testing.T) {
	detector := NewLooseImagesDetector()
	if detector.Name() != "loose-images" {
		t.Errorf("Expected name 'loose-images', got %q", detector.Name())
	}
}

func TestLooseImagesDetector_Detect(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create retina variant images
	testutil.CreateTestFile(t, tempDir, "icon@1x.png", 1000)
	testutil.CreateTestFile(t, tempDir, "icon@2x.png", 4000)
	testutil.CreateTestFile(t, tempDir, "icon@3x.png", 9000)

	detector := NewLooseImagesDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(optimizations) == 0 {
		t.Error("Expected optimizations for retina variants, got none")
	}

	// Verify optimization structure
	for _, opt := range optimizations {
		if opt.Category != "loose-images" {
			t.Errorf("Expected category 'loose-images', got %q", opt.Category)
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

func TestLooseImagesDetector_Detect_NoPatterns(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create single images (no patterns)
	testutil.CreatePNGFile(t, tempDir, "icon.png")

	detector := NewLooseImagesDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// Should not generate optimizations for single images without patterns
	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for single images, got %d", len(optimizations))
	}
}

func TestLooseImagesDetector_Detect_EmptyDirectory(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	detector := NewLooseImagesDetector()

	optimizations, err := detector.Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	if len(optimizations) != 0 {
		t.Errorf("Expected no optimizations for empty directory, got %d", len(optimizations))
	}
}

func TestDetectLooseImages_InAssetCatalog(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create asset catalog
	catalogDir := testutil.CreateTestDir(t, tempDir, "Assets.xcassets")
	testutil.CreatePNGFile(t, catalogDir, "AppIcon.png")

	looseImages, err := DetectLooseImages(tempDir)
	if err != nil {
		t.Fatalf("DetectLooseImages failed: %v", err)
	}

	// Should not find images inside asset catalogs
	if len(looseImages) != 0 {
		t.Errorf("Expected no loose images in asset catalog, got %d", len(looseImages))
		for _, img := range looseImages {
			t.Logf("Found: %s", img.Path)
		}
	}
}
