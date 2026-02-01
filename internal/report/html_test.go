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

	// Verify artifact info is displayed
	if !strings.Contains(output, "test.ipa") {
		t.Error("Output missing artifact name")
	}
	if !strings.Contains(output, "ipa") {
		t.Error("Output missing artifact type")
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
		expected string
	}{
		{"test.framework", "framework"},
		{"libtest.dylib", "library"},
		{"libtest.a", "library"},
		{"libtest.so", "native"},
		{"image.png", "image"},
		{"image.jpg", "image"},
		{"Assets.car", "asset_catalog"},
		{"Info.plist", "resource"},
		{"Main.storyboard", "ui"},
		{"classes.dex", "dex"},
		{"unknown.xyz", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := formatter.getFileType(tt.filename)
			if result != tt.expected {
				t.Errorf("getFileType(%q) = %q, want %q", tt.filename, result, tt.expected)
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
