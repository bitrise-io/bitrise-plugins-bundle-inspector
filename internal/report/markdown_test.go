package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestMarkdownFormatter_Format_EmptyReport(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path: "/path/to/empty.ipa",
			Type: "ipa",
			Size: 0,
		},
		SizeBreakdown: types.SizeBreakdown{},
		FileTree:      []*types.FileNode{},
		LargestFiles:  []types.FileNode{},
		Duplicates:    []types.DuplicateSet{},
		Optimizations: []types.Optimization{},
		TotalSavings:  0,
	}

	var buf bytes.Buffer
	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Check header exists
	if !strings.Contains(output, "## Bitrise Report") {
		t.Error("Output missing header")
	}

	// Check artifact info
	if !strings.Contains(output, "empty.ipa") {
		t.Error("Output missing artifact name")
	}

	// Check summary table exists
	if !strings.Contains(output, "| Bundle | Commit | Install Size | Download Size | Potential Savings |") {
		t.Error("Output missing summary table header")
	}
}

func TestMarkdownFormatter_Format_CompleteReport(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := createTestReport()

	var buf bytes.Buffer
	err := formatter.Format(&buf, report)
	if err != nil {
		t.Fatalf("Format() failed: %v", err)
	}

	output := buf.String()

	// Check all major sections exist
	expectedSections := []string{
		"## Bitrise Report",
		"| Bundle | Commit | Install Size | Download Size | Potential Savings |",
		"🔧 Strip Binary Symbols",
		"📊 Size Breakdown by Category",
		"📦 Top",
		"🔄 Duplicate Files Found",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output missing section: %s", section)
		}
	}

	// Check HTML tags are balanced
	detailsOpen := strings.Count(output, "<details")
	detailsClose := strings.Count(output, "</details>")
	if detailsOpen != detailsClose {
		t.Errorf("Unbalanced <details> tags: %d open, %d close", detailsOpen, detailsClose)
	}

	summaryOpen := strings.Count(output, "<summary>")
	summaryClose := strings.Count(output, "</summary>")
	if summaryOpen != summaryClose {
		t.Errorf("Unbalanced <summary> tags: %d open, %d close", summaryOpen, summaryClose)
	}
}

func TestMarkdownFormatter_writeHeader(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path: "/path/to/TestApp.ipa",
			Type: "ipa",
			Size: 10 * 1024 * 1024, // 10 MB
		},
		SizeBreakdown: types.SizeBreakdown{
			Frameworks: 5 * 1024 * 1024,
			Resources:  3 * 1024 * 1024,
			Executable: 2 * 1024 * 1024,
		},
		FileTree: []*types.FileNode{
			{Path: "file1.txt", IsDir: false},
			{Path: "file2.txt", IsDir: false},
		},
		Optimizations: []types.Optimization{
			{Category: "strip-symbols", Severity: "high", Impact: 1024},
			{Category: "duplicates", Severity: "medium", Impact: 512},
		},
		TotalSavings: 1536,
	}

	var buf bytes.Buffer
	err := formatter.writeHeader(&buf, report)
	if err != nil {
		t.Fatalf("writeHeader() failed: %v", err)
	}

	output := buf.String()

	// Check key elements
	if !strings.Contains(output, "TestApp.ipa") {
		t.Error("Missing artifact name")
	}
	if !strings.Contains(output, "## Bitrise Report") {
		t.Error("Missing Bitrise Report header")
	}
	if !strings.Contains(output, "| Bundle | Commit | Install Size | Download Size | Potential Savings |") {
		t.Error("Missing table header")
	}
	if !strings.Contains(output, "10.0 MB") {
		t.Error("Missing size information")
	}
	if !strings.Contains(output, "**Analysis:**") {
		t.Error("Missing analysis summary")
	}
	if !strings.Contains(output, "2 optimizations found across 2 categories") {
		t.Error("Missing optimization counts")
	}
}

func TestMarkdownFormatter_writeSizeBreakdown(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		SizeBreakdown: types.SizeBreakdown{
			Frameworks: 5 * 1024 * 1024,
			Resources:  3 * 1024 * 1024,
			Executable: 2 * 1024 * 1024,
		},
	}

	var buf bytes.Buffer
	err := formatter.writeSizeBreakdown(&buf, report)
	if err != nil {
		t.Fatalf("writeSizeBreakdown() failed: %v", err)
	}

	output := buf.String()

	// Check section exists
	if !strings.Contains(output, "📊 Size Breakdown by Category") {
		t.Error("Missing section title")
	}

	// Check table structure
	if !strings.Contains(output, "| Category | Size | Percentage |") {
		t.Error("Missing table header")
	}

	// Check categories appear (largest first)
	if !strings.Contains(output, "Frameworks") {
		t.Error("Missing Frameworks category")
	}
	if !strings.Contains(output, "Resources") {
		t.Error("Missing Resources category")
	}

	// Check HTML tags
	if !strings.Contains(output, "<details>") || !strings.Contains(output, "</details>") {
		t.Error("Missing details tags")
	}
}

func TestMarkdownFormatter_writeSizeBreakdown_Empty(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		SizeBreakdown: types.SizeBreakdown{},
	}

	var buf bytes.Buffer
	err := formatter.writeSizeBreakdown(&buf, report)
	if err != nil {
		t.Fatalf("writeSizeBreakdown() failed: %v", err)
	}

	output := buf.String()

	// Should produce no output for empty breakdown
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

func TestMarkdownFormatter_writeLargestFiles(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		LargestFiles: []types.FileNode{
			{Path: "Frameworks/BigSDK.framework/BigSDK", Size: 8 * 1024 * 1024},
			{Path: "Assets.car", Size: 5 * 1024 * 1024},
			{Path: "Resources/video.mp4", Size: 3 * 1024 * 1024},
		},
		SizeBreakdown: types.SizeBreakdown{
			Frameworks: 10 * 1024 * 1024,
			Resources:  6 * 1024 * 1024,
		},
	}

	var buf bytes.Buffer
	err := formatter.writeLargestFiles(&buf, report)
	if err != nil {
		t.Fatalf("writeLargestFiles() failed: %v", err)
	}

	output := buf.String()

	// Check section exists
	if !strings.Contains(output, "📦 Top") {
		t.Error("Missing section title")
	}

	// Check files appear
	if !strings.Contains(output, "BigSDK") {
		t.Error("Missing largest file")
	}
	if !strings.Contains(output, "Assets.car") {
		t.Error("Missing second file")
	}

	// Check numbering
	if !strings.Contains(output, "1. `") {
		t.Error("Missing numbered list")
	}
}

func TestMarkdownFormatter_writeLargestFiles_Empty(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		LargestFiles: []types.FileNode{},
	}

	var buf bytes.Buffer
	err := formatter.writeLargestFiles(&buf, report)
	if err != nil {
		t.Fatalf("writeLargestFiles() failed: %v", err)
	}

	output := buf.String()

	// Should produce no output for empty list
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

func TestMarkdownFormatter_writeDuplicates_WithDuplicates(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		Duplicates: []types.DuplicateSet{
			{
				Hash:  "abc123",
				Size:  1024,
				Count: 3,
				Files: []string{
					"Assets/icon.png",
					"Backup/icon.png",
					"Resources/icon.png",
				},
				WastedSize: 2048, // (3-1) * 1024
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.writeDuplicates(&buf, report)
	if err != nil {
		t.Fatalf("writeDuplicates() failed: %v", err)
	}

	output := buf.String()

	// Check section exists
	if !strings.Contains(output, "🔄 Duplicate Files Found") {
		t.Error("Missing section title")
	}

	// Check duplicate info
	if !strings.Contains(output, "3 copies of") {
		t.Error("Missing copy count")
	}
	if !strings.Contains(output, "`icon.png`") {
		t.Error("Missing file name")
	}
	if !strings.Contains(output, "**Wasted:**") {
		t.Error("Missing wasted space")
	}

	// Check all paths appear
	for _, path := range report.Duplicates[0].Files {
		if !strings.Contains(output, path) {
			t.Errorf("Missing path: %s", path)
		}
	}
}

func TestMarkdownFormatter_writeDuplicates_NoDuplicates(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		Duplicates: []types.DuplicateSet{},
	}

	var buf bytes.Buffer
	err := formatter.writeDuplicates(&buf, report)
	if err != nil {
		t.Fatalf("writeDuplicates() failed: %v", err)
	}

	output := buf.String()

	// Should produce no output when no duplicates
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

func TestMarkdownFormatter_writeOptimizations_Category(t *testing.T) {
	formatter := NewMarkdownFormatter()
	opts := []types.Optimization{
		{
			Category:    "strip-symbols",
			Title:       "Strip debug symbols from binary",
			Description: "Binary contains debug symbols",
			Severity:    "high",
			Impact:      1024 * 1024,
			Action:      "Run strip -x on binary",
			Files:       []string{"Frameworks/SDK.framework/SDK"},
		},
	}

	var buf bytes.Buffer
	err := formatter.writeOptimizations(&buf, opts, "Strip Binary Symbols", "🔧", true)
	if err != nil {
		t.Fatalf("writeOptimizations() failed: %v", err)
	}

	output := buf.String()

	// Check section is open
	if !strings.Contains(output, "<details open>") {
		t.Error("First category should be open by default")
	}

	// Check category name
	if !strings.Contains(output, "Strip Binary Symbols") {
		t.Error("Missing category name")
	}

	// Check optimization details
	if !strings.Contains(output, "Strip debug symbols from binary") {
		t.Error("Missing optimization title")
	}
	if !strings.Contains(output, "**Impact:**") {
		t.Error("Missing impact")
	}
	if !strings.Contains(output, "**Action:**") {
		t.Error("Missing action")
	}
	if !strings.Contains(output, "**Files:**") {
		t.Error("Missing files section")
	}
}

func TestMarkdownFormatter_writeOptimizations_Empty(t *testing.T) {
	formatter := NewMarkdownFormatter()
	opts := []types.Optimization{}

	var buf bytes.Buffer
	err := formatter.writeOptimizations(&buf, opts, "Strip Binary Symbols", "🔧", true)
	if err != nil {
		t.Fatalf("writeOptimizations() failed: %v", err)
	}

	output := buf.String()

	// Check positive message appears
	if !strings.Contains(output, "✅ No issues found!") {
		t.Error("Missing positive message for empty optimizations")
	}
}

func TestMarkdownFormatter_writeByExtension(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		SizeBreakdown: types.SizeBreakdown{
			ByExtension: map[string]int64{
				".png":   8 * 1024 * 1024,
				".dylib": 6 * 1024 * 1024,
				".ttf":   2 * 1024 * 1024,
			},
			Frameworks: 10 * 1024 * 1024,
			Resources:  6 * 1024 * 1024,
		},
	}

	var buf bytes.Buffer
	err := formatter.writeByExtension(&buf, report)
	if err != nil {
		t.Fatalf("writeByExtension() failed: %v", err)
	}

	output := buf.String()

	// Check section exists
	if !strings.Contains(output, "🔍 Size by File Extension") {
		t.Error("Missing section title")
	}

	// Check table
	if !strings.Contains(output, "| Extension | Size | Percentage |") {
		t.Error("Missing table header")
	}

	// Check extensions appear
	if !strings.Contains(output, ".png") {
		t.Error("Missing .png extension")
	}
	if !strings.Contains(output, ".dylib") {
		t.Error("Missing .dylib extension")
	}
}

func TestMarkdownFormatter_writeByExtension_Empty(t *testing.T) {
	formatter := NewMarkdownFormatter()
	report := &types.Report{
		SizeBreakdown: types.SizeBreakdown{
			ByExtension: map[string]int64{},
		},
	}

	var buf bytes.Buffer
	err := formatter.writeByExtension(&buf, report)
	if err != nil {
		t.Fatalf("writeByExtension() failed: %v", err)
	}

	output := buf.String()

	// Should produce no output for empty map
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		maxLen int
		want   string
	}{
		{
			name:   "short path unchanged",
			path:   "short/path.txt",
			maxLen: 50,
			want:   "short/path.txt",
		},
		{
			name:   "long path truncated",
			path:   "very/long/path/that/exceeds/maximum/length/file.txt",
			maxLen: 30,
			want:   "very/long/pa/.../ile.txt",
		},
		{
			name:   "exact length unchanged",
			path:   "exact/path.txt",
			maxLen: 14,
			want:   "exact/path.txt",
		},
		{
			name:   "very small maxLen",
			path:   "path/file.txt",
			maxLen: 10,
			want:   "pa/.../xt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePath(tt.path, tt.maxLen)
			if len(got) > tt.maxLen {
				t.Errorf("truncatePath() length %d exceeds max %d", len(got), tt.maxLen)
			}
			// For short paths, should be unchanged
			if len(tt.path) <= tt.maxLen && got != tt.want {
				t.Errorf("truncatePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetCategoryGroups(t *testing.T) {
	opts := []types.Optimization{
		{Category: "strip-symbols", Impact: 1000},
		{Category: "strip-symbols", Impact: 2000},
		{Category: "duplicates", Impact: 500},
		{Category: "image-optimization", Impact: 600},
		{Category: "loose-images", Impact: 100},
	}

	groups := getCategoryGroups(opts)

	// Check all categories present
	if len(groups["strip-symbols"]) != 2 {
		t.Errorf("Expected 2 strip-symbols, got %d", len(groups["strip-symbols"]))
	}
	if len(groups["duplicates"]) != 1 {
		t.Errorf("Expected 1 duplicates, got %d", len(groups["duplicates"]))
	}
	if len(groups["image-optimization"]) != 1 {
		t.Errorf("Expected 1 image-optimization, got %d", len(groups["image-optimization"]))
	}
	if len(groups["loose-images"]) != 1 {
		t.Errorf("Expected 1 loose-images, got %d", len(groups["loose-images"]))
	}
}

func TestCalculateSavings(t *testing.T) {
	opts := []types.Optimization{
		{Impact: 1000},
		{Impact: 2000},
		{Impact: 500},
	}

	total := calculateSavings(opts)
	expected := int64(3500)

	if total != expected {
		t.Errorf("calculateSavings() = %d, want %d", total, expected)
	}
}

func TestCalculateSavings_Empty(t *testing.T) {
	opts := []types.Optimization{}
	total := calculateSavings(opts)

	if total != 0 {
		t.Errorf("calculateSavings() = %d, want 0", total)
	}
}

func TestSortBySize(t *testing.T) {
	breakdown := map[string]int64{
		"Small":  100,
		"Large":  1000,
		"Medium": 500,
	}

	sorted := sortBySize(breakdown)

	// Should be sorted descending
	if len(sorted) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(sorted))
	}

	if sorted[0].name != "Large" || sorted[0].size != 1000 {
		t.Errorf("First item should be Large:1000, got %s:%d", sorted[0].name, sorted[0].size)
	}
	if sorted[1].name != "Medium" || sorted[1].size != 500 {
		t.Errorf("Second item should be Medium:500, got %s:%d", sorted[1].name, sorted[1].size)
	}
	if sorted[2].name != "Small" || sorted[2].size != 100 {
		t.Errorf("Third item should be Small:100, got %s:%d", sorted[2].name, sorted[2].size)
	}
}

func TestCountFiles(t *testing.T) {
	nodes := []*types.FileNode{
		{Path: "dir1", IsDir: true, Children: []*types.FileNode{
			{Path: "dir1/file1.txt", IsDir: false},
			{Path: "dir1/file2.txt", IsDir: false},
		}},
		{Path: "file3.txt", IsDir: false},
		{Path: "dir2", IsDir: true, Children: []*types.FileNode{
			{Path: "dir2/subdir", IsDir: true, Children: []*types.FileNode{
				{Path: "dir2/subdir/file4.txt", IsDir: false},
			}},
		}},
	}

	count := countFiles(nodes)
	expected := 4 // file1, file2, file3, file4

	if count != expected {
		t.Errorf("countFiles() = %d, want %d", count, expected)
	}
}

func TestCalculateUncompressedSize(t *testing.T) {
	breakdown := &types.SizeBreakdown{
		Executable: 1000,
		Frameworks: 2000,
		Resources:  3000,
		Assets:     4000,
		Libraries:  500,
		DEX:        600,
		Other:      700,
	}

	total := calculateUncompressedSize(breakdown)
	expected := int64(11800)

	if total != expected {
		t.Errorf("calculateUncompressedSize() = %d, want %d", total, expected)
	}
}

func TestFindLargestCategory(t *testing.T) {
	breakdown := &types.SizeBreakdown{
		Executable: 1000,
		Frameworks: 5000,
		Resources:  3000,
		Assets:     2000,
	}

	name, size := findLargestCategory(breakdown)

	if name != "Frameworks" {
		t.Errorf("findLargestCategory() name = %s, want Frameworks", name)
	}
	if size != 5000 {
		t.Errorf("findLargestCategory() size = %d, want 5000", size)
	}
}

// Helper function to create a complete test report
func createTestReport() *types.Report {
	return &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Path: "/path/to/TestApp.ipa",
			Type: "ipa",
			Size: 45 * 1024 * 1024,
		},
		SizeBreakdown: types.SizeBreakdown{
			Frameworks: 22 * 1024 * 1024,
			Resources:  15 * 1024 * 1024,
			Executable: 8 * 1024 * 1024,
			Assets:     7 * 1024 * 1024,
			ByExtension: map[string]int64{
				".png":   8 * 1024 * 1024,
				".dylib": 6 * 1024 * 1024,
			},
		},
		FileTree: []*types.FileNode{
			{Path: "file1.txt", IsDir: false},
			{Path: "dir1", IsDir: true, Children: []*types.FileNode{
				{Path: "dir1/file2.txt", IsDir: false},
			}},
		},
		LargestFiles: []types.FileNode{
			{Path: "Frameworks/BigSDK.framework/BigSDK", Size: 8 * 1024 * 1024},
			{Path: "Assets.car", Size: 5 * 1024 * 1024},
		},
		Duplicates: []types.DuplicateSet{
			{
				Hash:       "abc123",
				Size:       120 * 1024,
				Count:      2,
				Files:      []string{"Assets/icon.png", "Backup/icon.png"},
				WastedSize: 120 * 1024,
			},
		},
		Optimizations: []types.Optimization{
			{
				Category:    "strip-symbols",
				Title:       "Strip debug symbols from WMF",
				Description: "Binary contains debug symbols",
				Severity:    "high",
				Impact:      1800 * 1024,
				Action:      "Run strip -x on binary",
				Files:       []string{"Frameworks/WMF.framework/WMF"},
			},
			{
				Category: "duplicates",
				Title:    "Duplicate resource files",
				Severity: "medium",
				Impact:   500 * 1024,
			},
			{
				Category: "image-optimization",
				Title:    "Optimize PNG images",
				Severity: "medium",
				Impact:   100 * 1024,
			},
		},
		TotalSavings: 2400 * 1024,
	}
}
