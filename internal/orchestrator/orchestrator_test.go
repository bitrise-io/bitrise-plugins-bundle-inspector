package orchestrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name      string
		impact    int64
		totalSize int64
		want      string
	}{
		{
			name:      "high severity - 15% impact",
			impact:    15 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "high",
		},
		{
			name:      "high severity - exactly 10% impact",
			impact:    10 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "high",
		},
		{
			name:      "medium severity - 7% impact",
			impact:    7 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "medium",
		},
		{
			name:      "medium severity - exactly 5% impact",
			impact:    5 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "medium",
		},
		{
			name:      "low severity - 2% impact",
			impact:    2 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "low",
		},
		{
			name:      "low severity - zero total",
			impact:    1 * 1024 * 1024,
			totalSize: 0,
			want:      "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSeverity(tt.impact, tt.totalSize)
			if got != tt.want {
				t.Errorf("getSeverity(%d, %d) = %s, want %s", tt.impact, tt.totalSize, got, tt.want)
			}
		})
	}
}

func TestCalculateTotalSavings(t *testing.T) {
	report := &types.Report{
		Optimizations: []types.Optimization{
			{Impact: 1024 * 1024},     // 1 MB
			{Impact: 2 * 1024 * 1024}, // 2 MB
			{Impact: 500 * 1024},      // 500 KB
			{Impact: 0},               // 0 bytes
		},
	}

	expected := int64(1024*1024 + 2*1024*1024 + 500*1024)
	got := calculateTotalSavings(report)

	if got != expected {
		t.Errorf("calculateTotalSavings() = %d, want %d", got, expected)
	}
}

func TestCalculateTotalSavings_EmptyOptimizations(t *testing.T) {
	report := &types.Report{
		Optimizations: []types.Optimization{},
	}

	expected := int64(0)
	got := calculateTotalSavings(report)

	if got != expected {
		t.Errorf("calculateTotalSavings() with empty optimizations = %d, want %d", got, expected)
	}
}

func TestNew(t *testing.T) {
	orch := New()

	if orch == nil {
		t.Fatal("New() returned nil")
	}

	if !orch.IncludeDuplicates {
		t.Error("Expected IncludeDuplicates to be true by default")
	}
}

func TestAnnotateFileTreeDuplicates(t *testing.T) {
	orch := New()
	hash := "abc123def456"

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{Type: types.ArtifactTypeAPK},
		Duplicates: []types.DuplicateSet{
			{
				Hash:  hash,
				Size:  4096,
				Count: 2,
				Files: []string{"res/icon.png", "res/backup/icon.png"},
			},
		},
		FileTree: []*types.FileNode{
			{
				Path: "res", Name: "res", IsDir: true,
				Children: []*types.FileNode{
					{Path: "res/icon.png", Name: "icon.png", Size: 1024},
					{Path: "res/layout.xml", Name: "layout.xml", Size: 512},
					{
						Path: "res/backup", Name: "backup", IsDir: true,
						Children: []*types.FileNode{
							{Path: "res/backup/icon.png", Name: "icon.png", Size: 1024},
						},
					},
				},
			},
		},
	}

	// Use empty extractPath for APK (no prefix needed)
	orch.annotateFileTreeDuplicates(report, "/tmp/extract")

	// Check annotated nodes
	iconNode := report.FileTree[0].Children[0]
	if !iconNode.IsDuplicate {
		t.Error("Expected res/icon.png to be marked as duplicate")
	}
	if iconNode.Hash != hash {
		t.Errorf("Expected hash %s, got %s", hash, iconNode.Hash)
	}

	// Check non-duplicate node is NOT annotated
	layoutNode := report.FileTree[0].Children[1]
	if layoutNode.IsDuplicate {
		t.Error("Expected res/layout.xml to NOT be marked as duplicate")
	}
	if layoutNode.Hash != "" {
		t.Error("Expected empty hash for non-duplicate node")
	}

	// Check nested duplicate
	backupIconNode := report.FileTree[0].Children[2].Children[0]
	if !backupIconNode.IsDuplicate {
		t.Error("Expected res/backup/icon.png to be marked as duplicate")
	}
	if backupIconNode.Hash != hash {
		t.Errorf("Expected hash %s, got %s", hash, backupIconNode.Hash)
	}
}

func TestAnnotateFileTreeDuplicates_NoDuplicates(t *testing.T) {
	orch := New()

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{Type: types.ArtifactTypeAPK},
		Duplicates:   nil,
		FileTree: []*types.FileNode{
			{Path: "file.txt", Name: "file.txt", Size: 100},
		},
	}

	// Should not panic or modify anything
	orch.annotateFileTreeDuplicates(report, "/tmp/extract")

	if report.FileTree[0].IsDuplicate {
		t.Error("Expected no annotation when there are no duplicates")
	}
}

func TestComputeDuplicatePathPrefix_IPA(t *testing.T) {
	orch := New()

	// Create temp dir simulating IPA extraction structure
	tempDir, err := os.MkdirTemp("", "prefix-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	appDir := filepath.Join(tempDir, "Payload", "MyApp.app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("Failed to create app dir: %v", err)
	}

	prefix := orch.computeDuplicatePathPrefix(types.ArtifactTypeIPA, tempDir)
	expected := "Payload/MyApp.app/"
	if prefix != expected {
		t.Errorf("Expected prefix %q, got %q", expected, prefix)
	}
}

func TestComputeDuplicatePathPrefix_APK(t *testing.T) {
	orch := New()

	prefix := orch.computeDuplicatePathPrefix(types.ArtifactTypeAPK, "/tmp/whatever")
	if prefix != "" {
		t.Errorf("Expected empty prefix for APK, got %q", prefix)
	}
}

func TestGenerateOptimizations_ExtensionDuplicationMessage(t *testing.T) {
	orch := New()

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{
			Size: 100 * 1024 * 1024,
		},
		Duplicates: []types.DuplicateSet{
			{
				// Extension duplication - should get specific message
				Hash:  "ext_hash",
				Size:  600 * 1024,
				Count: 2,
				Files: []string{
					"Payload/App.app/logo.png",
					"Payload/App.app/PlugIns/ShareExtension.appex/logo.png",
				},
				WastedSize: 600 * 1024,
			},
			{
				// Regular duplication - should get default message
				Hash:  "reg_hash",
				Size:  200 * 1024,
				Count: 2,
				Files: []string{
					"Payload/App.app/icon.png",
					"Payload/App.app/Resources/icon.png",
				},
				WastedSize: 200 * 1024,
			},
		},
	}

	optimizations := orch.generateOptimizations(report)

	// Find the extension duplication optimization
	var extOpt, regOpt *types.Optimization
	for i := range optimizations {
		for _, f := range optimizations[i].Files {
			if f == "Payload/App.app/PlugIns/ShareExtension.appex/logo.png" {
				extOpt = &optimizations[i]
			}
			if f == "Payload/App.app/Resources/icon.png" {
				regOpt = &optimizations[i]
			}
		}
	}

	if extOpt == nil {
		t.Fatal("Expected to find extension duplication optimization")
	}
	if extOpt.Action != "Consider using App Groups for shared data or an embedded framework to reduce duplication across extensions" {
		t.Errorf("Expected specific extension message, got: %s", extOpt.Action)
	}

	if regOpt == nil {
		t.Fatal("Expected to find regular duplication optimization")
	}
	if regOpt.Action != "Keep only one copy and deduplicate references" {
		t.Errorf("Expected default message, got: %s", regOpt.Action)
	}
}

func TestAnnotateFileTreeDuplicates_IPAPrefix(t *testing.T) {
	orch := New()
	hash := "sha256hash"

	// Create temp dir simulating IPA extraction
	tempDir, err := os.MkdirTemp("", "ipa-annotate-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	appDir := filepath.Join(tempDir, "Payload", "Test.app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("Failed to create app dir: %v", err)
	}

	report := &types.Report{
		ArtifactInfo: types.ArtifactInfo{Type: types.ArtifactTypeIPA},
		Duplicates: []types.DuplicateSet{
			{
				Hash:  hash,
				Size:  4096,
				Count: 2,
				// Duplicate paths are relative to extract root (include Payload/App.app/)
				Files: []string{
					"Payload/Test.app/Frameworks/A.framework/icon.png",
					"Payload/Test.app/Frameworks/B.framework/icon.png",
				},
			},
		},
		// FileNode paths are relative to .app root (no Payload/App.app/ prefix)
		FileTree: []*types.FileNode{
			{
				Path: "Frameworks", Name: "Frameworks", IsDir: true,
				Children: []*types.FileNode{
					{
						Path: "Frameworks/A.framework", Name: "A.framework", IsDir: true,
						Children: []*types.FileNode{
							{Path: "Frameworks/A.framework/icon.png", Name: "icon.png", Size: 1024},
						},
					},
					{
						Path: "Frameworks/B.framework", Name: "B.framework", IsDir: true,
						Children: []*types.FileNode{
							{Path: "Frameworks/B.framework/icon.png", Name: "icon.png", Size: 1024},
						},
					},
				},
			},
		},
	}

	orch.annotateFileTreeDuplicates(report, tempDir)

	// Both icons should be annotated despite the path coordinate difference
	iconA := report.FileTree[0].Children[0].Children[0]
	if !iconA.IsDuplicate || iconA.Hash != hash {
		t.Errorf("Expected A/icon.png annotated (IsDuplicate=%v, Hash=%s)", iconA.IsDuplicate, iconA.Hash)
	}

	iconB := report.FileTree[0].Children[1].Children[0]
	if !iconB.IsDuplicate || iconB.Hash != hash {
		t.Errorf("Expected B/icon.png annotated (IsDuplicate=%v, Hash=%s)", iconB.IsDuplicate, iconB.Hash)
	}
}
