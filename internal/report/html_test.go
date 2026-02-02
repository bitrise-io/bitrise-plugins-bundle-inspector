package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestHTMLFormatter_Format(t *testing.T) {
	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       10485760, // 10 MB
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{
			Executable:  5242880,
			Frameworks:  3145728,
			Resources:   1048576,
			Assets:      1048576,
			Libraries:   0,
			DEX:         0,
			Other:       0,
			ByExtension: map[string]int64{
				".png": 524288,
				".jpg": 262144,
			},
		},
		FileTree: []*types.FileNode{
			{
				Name:  "Payload",
				Path:  "Payload",
				Size:  10485760,
				IsDir: true,
				Children: []*types.FileNode{
					{
						Name:  "App.app",
						Path:  "Payload/App.app",
						Size:  10485760,
						IsDir: true,
						Children: []*types.FileNode{
							{
								Name:  "App",
								Path:  "Payload/App.app/App",
								Size:  5242880,
								IsDir: false,
							},
							{
								Name:  "Frameworks",
								Path:  "Payload/App.app/Frameworks",
								Size:  3145728,
								IsDir: true,
								Children: []*types.FileNode{
									{
										Name:  "Flutter.framework",
										Path:  "Payload/App.app/Frameworks/Flutter.framework",
										Size:  3145728,
										IsDir: false,
									},
								},
							},
						},
					},
				},
			},
		},
		Duplicates: []types.DuplicateSet{
			{
				Hash:       "abc123",
				Size:       1024,
				Count:      3,
				Files:      []string{"file1.png", "file2.png", "file3.png"},
				WastedSize: 2048,
			},
		},
		Optimizations: []types.Optimization{
			{
				Category:    "image-optimization",
				Severity:    "high",
				Title:       "Large uncompressed images",
				Description: "Found images that could be optimized",
				Impact:      524288,
				Files:       []string{"image1.png", "image2.jpg"},
				Action:      "Compress images using tools like ImageOptim",
			},
		},
		TotalSavings: 526336,
	}

	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify HTML structure
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output missing DOCTYPE declaration")
	}
	if !strings.Contains(output, "<title>") {
		t.Error("Output missing title tag")
	}
	if !strings.Contains(output, "echarts") {
		t.Error("Output missing ECharts reference")
	}

	// Verify data is embedded
	if !strings.Contains(output, "reportData") {
		t.Error("Output missing reportData variable")
	}
	if !strings.Contains(output, "fileTree") {
		t.Error("Output missing fileTree data")
	}
	if !strings.Contains(output, "categories") {
		t.Error("Output missing categories data")
	}

	// Verify artifact info is displayed (type is included in the title or data)
	if !strings.Contains(output, "Analysis Report") {
		t.Error("Output missing analysis report title")
	}

	// Verify visualizations are included
	if !strings.Contains(output, "createTreemap") {
		t.Error("Output missing treemap function")
	}
	if !strings.Contains(output, "createCategoryChart") {
		t.Error("Output missing category chart function")
	}
	if !strings.Contains(output, "createExtensionChart") {
		t.Error("Output missing extension chart function")
	}

	// Verify optimizations section
	if !strings.Contains(output, "Optimization Opportunities") {
		t.Error("Output missing optimizations section")
	}
}

func TestHTMLFormatter_EmptyReport(t *testing.T) {
	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/empty.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       0,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
		Duplicates:    []types.DuplicateSet{},
		Optimizations: []types.Optimization{},
		TotalSavings:  0,
	}

	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed on empty report: %v", err)
	}

	output := buf.String()

	// Should still produce valid HTML
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Empty report missing DOCTYPE declaration")
	}
}

func TestPrepareTreemapData(t *testing.T) {
	formatter := NewHTMLFormatter()
	nodes := []*types.FileNode{
		{
			Name:  "root",
			Path:  "/",
			Size:  1000,
			IsDir: true,
			Children: []*types.FileNode{
				{
					Name:  "file1.txt",
					Path:  "/file1.txt",
					Size:  500,
					IsDir: false,
				},
				{
					Name:  "file2.txt",
					Path:  "/file2.txt",
					Size:  500,
					IsDir: false,
				},
			},
		},
	}

	result := formatter.prepareTreemapData(nodes)

	// Verify result is a map
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("prepareTreemapData should return a map")
	}

	// Verify required fields
	if _, ok := resultMap["name"]; !ok {
		t.Error("Result missing 'name' field")
	}
	if _, ok := resultMap["value"]; !ok {
		t.Error("Result missing 'value' field")
	}
}

func TestGetFileType(t *testing.T) {
	formatter := NewHTMLFormatter()

	tests := []struct {
		filename string
		path     string
		expected string
	}{
		// Frameworks
		{"test.framework", "Payload/App.app/Frameworks/test.framework", "framework"},
		// Libraries
		{"libtest.dylib", "Payload/App.app/Frameworks/libtest.dylib", "library"},
		{"libtest.a", "Payload/App.app/libtest.a", "library"},
		{"libtest.so", "lib/arm64-v8a/libtest.so", "native"},
		// Images (including new types)
		{"image.png", "Resources/image.png", "image"},
		{"image.jpg", "Resources/image.jpg", "image"},
		{"image.gif", "Resources/image.gif", "image"},
		{"image.heic", "Resources/image.heic", "image"},
		{"image.heif", "Resources/image.heif", "image"},
		{"image.bmp", "Resources/image.bmp", "image"},
		{"image.tiff", "Resources/image.tiff", "image"},
		{"image.webp", "Resources/image.webp", "image"},
		{"image.svg", "Resources/image.svg", "image"},
		// Asset catalogs
		{"Assets.car", "Payload/App.app/Assets.car", "asset_catalog"},
		// Virtual paths under asset catalogs at any depth (NEW)
		{"AppIcon", "Payload/App.app/Assets.car/AppIcon", "asset_catalog"},
		{"logo@2x.png", "Payload/App.app/Assets.car/Images/logo@2x.png", "asset_catalog"},
		{"deep-icon.png", "Payload/App.app/Assets.car/Images/Icons/Deep/deep-icon.png", "asset_catalog"},
		// Resources
		{"Info.plist", "Payload/App.app/Info.plist", "resource"},
		{"config.xml", "res/config.xml", "resource"},
		{"data.json", "data/data.json", "resource"},
		// UI
		{"Main.storyboard", "Base.lproj/Main.storyboard", "ui"},
		{"View.xib", "Base.lproj/View.xib", "ui"},
		// Android
		{"classes.dex", "classes.dex", "dex"},
		// Fonts (including new types)
		{"font.ttf", "Fonts/font.ttf", "font"},
		{"font.otf", "Fonts/font.otf", "font"},
		{"font.woff", "Fonts/font.woff", "font"},
		{"font.woff2", "Fonts/font.woff2", "font"},
		// Video (new)
		{"video.mp4", "Resources/video.mp4", "video"},
		{"video.mov", "Resources/video.mov", "video"},
		{"video.m4v", "Resources/video.m4v", "video"},
		{"video.avi", "Resources/video.avi", "video"},
		// Audio (new)
		{"audio.mp3", "Resources/audio.mp3", "audio"},
		{"audio.m4a", "Resources/audio.m4a", "audio"},
		{"audio.wav", "Resources/audio.wav", "audio"},
		{"audio.aac", "Resources/audio.aac", "audio"},
		{"audio.caf", "Resources/audio.caf", "audio"},
		// ML Models (new)
		{"model.mlmodel", "Models/model.mlmodel", "mlmodel"},
		{"model.mlmodelc", "Models/model.mlmodelc", "mlmodel"},
		// Localization (new)
		{"Localizable.strings", "en.lproj/Localizable.strings", "localization"},
		{"Plurals.stringsdict", "en.lproj/Plurals.stringsdict", "localization"},
		{"en.lproj", "en.lproj", "localization"},
		// Binary virtual paths at any depth (NEW)
		{"[__TEXT]", "Payload/App.app/MyApp/[__TEXT]", "other"},
		{"[__DATA]", "Payload/App.app/libtest.dylib/[__DATA]", "library"},
		{"[__TEXT]", "Payload/App.app/Frameworks/Flutter.framework/Flutter/[__TEXT]", "framework"},
		// Deeply nested binary virtual paths (ANY DEPTH)
		{"subsection", "Payload/App.app/MyApp/[__TEXT]/subsection", "other"},
		{"deep-data", "Payload/App.app/libtest.dylib/[__DATA]/deep/deep-data", "library"},
		{"nested", "Payload/App.app/Frameworks/Flutter.framework/Flutter/[__TEXT]/code/nested", "framework"},
		{"native-section", "lib/arm64-v8a/libgame.so/[__TEXT]/native-section", "native"},
		// Other
		{"unknown.xyz", "unknown.xyz", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := formatter.getFileType(tt.filename, tt.path)
			if result != tt.expected {
				t.Errorf("getFileType(%q, %q) = %q, want %q", tt.filename, tt.path, result, tt.expected)
			}
		})
	}
}

func TestPrepareCategoryData(t *testing.T) {
	formatter := NewHTMLFormatter()
	breakdown := &types.SizeBreakdown{
		Executable: 5000,
		Frameworks: 3000,
		Resources:  2000,
		Assets:     1000,
		Libraries:  500,
		DEX:        0,
		Other:      100,
	}

	result := formatter.prepareCategoryData(breakdown)

	// Should only include non-zero categories
	for _, cat := range result {
		if cat.Value == 0 {
			t.Errorf("Category %s has zero value, should be filtered out", cat.Name)
		}
	}

	// Should be sorted by size descending
	for i := 1; i < len(result); i++ {
		if result[i].Value > result[i-1].Value {
			t.Error("Categories not sorted by size descending")
		}
	}
}

func TestPrepareExtensionData(t *testing.T) {
	formatter := NewHTMLFormatter()
	breakdown := &types.SizeBreakdown{
		ByExtension: map[string]int64{
			".png":  5000,
			".jpg":  3000,
			".json": 2000,
			".txt":  1000,
		},
	}

	result := formatter.prepareExtensionData(breakdown)

	// Should be sorted by size descending
	for i := 1; i < len(result); i++ {
		if result[i].Value > result[i-1].Value {
			t.Error("Extensions not sorted by size descending")
		}
	}

	// Should limit to top 10 (or less)
	if len(result) > 10 {
		t.Error("Result should be limited to top 10 extensions")
	}
}

func TestCalculateNodeCount(t *testing.T) {
	formatter := NewHTMLFormatter()
	nodes := []*types.FileNode{
		{
			Name:  "root",
			Children: []*types.FileNode{
				{Name: "child1"},
				{
					Name: "child2",
					Children: []*types.FileNode{
						{Name: "grandchild1"},
						{Name: "grandchild2"},
					},
				},
			},
		},
	}

	count := formatter.calculateNodeCount(nodes)
	expected := 5 // root + child1 + child2 + grandchild1 + grandchild2
	if count != expected {
		t.Errorf("calculateNodeCount() = %d, want %d", count, expected)
	}
}

// Tests for security-related functionality

func TestSafeHTMLEscaping(t *testing.T) {
	// Verify that SafeHTML.escapeText is present in the template
	// and that the template properly escapes user-controlled data
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/<script>alert('xss')</script>.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify SafeHTML utility is included
	if !strings.Contains(output, "SafeHTML") {
		t.Error("Output missing SafeHTML utility")
	}

	// Verify SafeHTML.escapeText function is present
	if !strings.Contains(output, "escapeText") {
		t.Error("Output missing escapeText function")
	}
}

func TestURLValidation(t *testing.T) {
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify URL validation function is present
	if !strings.Contains(output, "isValidLearnMoreURL") {
		t.Error("Output missing isValidLearnMoreURL function")
	}

	// Verify allowed domains are defined
	if !strings.Contains(output, "ALLOWED_URL_DOMAINS") {
		t.Error("Output missing ALLOWED_URL_DOMAINS")
	}

	// Verify bitrise.io is in allowed domains
	if !strings.Contains(output, "devcenter.bitrise.io") {
		t.Error("Output missing devcenter.bitrise.io in allowed domains")
	}
}

func TestSafeGetElement(t *testing.T) {
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify safeGetElement function is present
	if !strings.Contains(output, "safeGetElement") {
		t.Error("Output missing safeGetElement function")
	}

	// Verify it's used instead of direct getElementById calls in critical places
	if strings.Count(output, "safeGetElement") < 5 {
		t.Error("Expected safeGetElement to be used multiple times")
	}
}

func TestEventDelegation(t *testing.T) {
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify event delegation is present
	if !strings.Contains(output, "setupEventDelegation") {
		t.Error("Output missing setupEventDelegation function")
	}

	// Verify data-action attributes are used
	if !strings.Contains(output, "data-action") {
		t.Error("Output missing data-action attributes")
	}
}

func TestAppStateConsolidation(t *testing.T) {
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify AppState object is present
	if !strings.Contains(output, "const AppState = {") {
		t.Error("Output missing AppState object")
	}

	// Verify AppState has key properties
	if !strings.Contains(output, "isDark()") {
		t.Error("Output missing AppState.isDark method")
	}
	if !strings.Contains(output, "setTheme") {
		t.Error("Output missing AppState.setTheme method")
	}
}

func TestChartFactoryPresence(t *testing.T) {
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify ChartFactory object is present
	if !strings.Contains(output, "const ChartFactory = {") {
		t.Error("Output missing ChartFactory object")
	}

	// Verify ChartFactory has unified resize handler
	if !strings.Contains(output, "resizeAll") {
		t.Error("Output missing ChartFactory.resizeAll method")
	}
}

func TestTreeUtilsPresence(t *testing.T) {
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify TreeUtils object is present
	if !strings.Contains(output, "const TreeUtils = {") {
		t.Error("Output missing TreeUtils object")
	}

	// Verify TreeUtils has key methods
	if !strings.Contains(output, "deepCopy") {
		t.Error("Output missing TreeUtils.deepCopy method")
	}
	if !strings.Contains(output, "traverse") {
		t.Error("Output missing TreeUtils.traverse method")
	}
	if !strings.Contains(output, "filter") {
		t.Error("Output missing TreeUtils.filter method")
	}
}

func TestAccessibilityFeatures(t *testing.T) {
	formatter := NewHTMLFormatter()
	var buf bytes.Buffer

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path:       "/path/to/test.ipa",
			Type:       types.ArtifactTypeIPA,
			Size:       1000,
			AnalyzedAt: time.Now(),
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
	}

	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Verify screen reader only class exists
	if !strings.Contains(output, "sr-only") {
		t.Error("Output missing sr-only CSS class")
	}

	// Verify search input has label
	if !strings.Contains(output, "for=\"search-input\"") {
		t.Error("Output missing label for search input")
	}

	// Verify aria attributes are used
	if !strings.Contains(output, "aria-label") {
		t.Error("Output missing aria-label attributes")
	}

	// Verify role attributes are used
	if !strings.Contains(output, "role=\"list\"") {
		t.Error("Output missing role=\"list\" attributes")
	}

	// Verify sections have aria-labelledby
	if !strings.Contains(output, "aria-labelledby") {
		t.Error("Output missing aria-labelledby attributes")
	}
}
