package android

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/testutil"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestAABAnalyzer_ValidateArtifact(t *testing.T) {
	tmpDir := testutil.CreateTempDir(t)

	tests := []struct {
		name      string
		setupPath func() string
		wantErr   bool
	}{
		{
			name: "valid AAB file",
			setupPath: func() string {
				return testutil.CreateTestFile(t, tmpDir, "test.aab", 100)
			},
			wantErr: false,
		},
		{
			name: "invalid extension",
			setupPath: func() string {
				return testutil.CreateTestFile(t, tmpDir, "test.apk", 100)
			},
			wantErr: true,
		},
		{
			name: "nonexistent file",
			setupPath: func() string {
				return filepath.Join(tmpDir, "nonexistent.aab")
			},
			wantErr: true,
		},
		{
			name: "directory instead of file",
			setupPath: func() string {
				dir := filepath.Join(tmpDir, "test_dir.aab")
				if err := os.Mkdir(dir, 0755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantErr: true,
		},
		{
			name: "case insensitive extension",
			setupPath: func() string {
				return testutil.CreateTestFile(t, tmpDir, "test_upper.AAB", 100)
			},
			wantErr: false,
		},
	}

	analyzer := NewAABAnalyzer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupPath()
			err := analyzer.ValidateArtifact(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateArtifact() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAABAnalyzer_Analyze(t *testing.T) {
	tmpDir := testutil.CreateTempDir(t)

	// Create a mock AAB file (ZIP with AAB structure)
	aabPath := filepath.Join(tmpDir, "test.aab")
	if err := createMockAAB(aabPath); err != nil {
		t.Fatalf("Failed to create mock AAB: %v", err)
	}

	analyzer := NewAABAnalyzer()
	ctx := context.Background()

	report, err := analyzer.Analyze(ctx, aabPath)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Validate report structure
	if report == nil {
		t.Fatal("Expected report, got nil")
	}

	if report.ArtifactInfo.Type != types.ArtifactTypeAAB {
		t.Errorf("ArtifactType = %v, want %v", report.ArtifactInfo.Type, types.ArtifactTypeAAB)
	}

	if report.ArtifactInfo.Path != aabPath {
		t.Errorf("Path = %v, want %v", report.ArtifactInfo.Path, aabPath)
	}

	if report.ArtifactInfo.Size == 0 {
		t.Error("Expected non-zero compressed size")
	}

	if report.ArtifactInfo.UncompressedSize == 0 {
		t.Error("Expected non-zero uncompressed size")
	}

	if report.FileTree == nil {
		t.Error("Expected file tree, got nil")
	}

	if report.SizeBreakdown.ByCategory == nil {
		t.Error("Expected size breakdown by category, got nil")
	}

	if report.SizeBreakdown.ByExtension == nil {
		t.Error("Expected size breakdown by extension, got nil")
	}

	if len(report.LargestFiles) == 0 {
		t.Error("Expected largest files list to be non-empty")
	}

	// Check metadata for modules
	if report.Metadata == nil {
		t.Error("Expected metadata, got nil")
	}

	modules, ok := report.Metadata["modules"]
	if !ok {
		t.Error("Expected modules in metadata")
	}

	modulesList, ok := modules.([]string)
	if !ok {
		t.Error("Expected modules to be []string")
	}

	if len(modulesList) == 0 {
		t.Error("Expected at least one module")
	}
}

func TestAABAnalyzer_Analyze_InvalidFile(t *testing.T) {
	tmpDir := testutil.CreateTempDir(t)

	// Create a non-ZIP file with .aab extension
	aabPath := testutil.CreateTestFile(t, tmpDir, "invalid.aab", 100)

	analyzer := NewAABAnalyzer()
	ctx := context.Background()

	_, err := analyzer.Analyze(ctx, aabPath)
	if err == nil {
		t.Error("Expected error for invalid AAB file, got nil")
	}
}

func TestDetectModules(t *testing.T) {
	tests := []struct {
		name        string
		fileTree    []*types.FileNode
		wantModules []string
	}{
		{
			name: "single base module",
			fileTree: []*types.FileNode{
				{Name: "base", Size: 1000, IsDir: true},
			},
			wantModules: []string{"base"},
		},
		{
			name: "multiple modules",
			fileTree: []*types.FileNode{
				{Name: "base", Size: 1000, IsDir: true},
				{Name: "feature1", Size: 500, IsDir: true},
				{Name: "feature2", Size: 500, IsDir: true},
			},
			wantModules: []string{"base", "feature1", "feature2"},
		},
		{
			name: "modules with files",
			fileTree: []*types.FileNode{
				{Name: "base", Size: 1000, IsDir: true},
				{Name: "BundleConfig.pb", Size: 100, IsDir: false},
			},
			wantModules: []string{"base"},
		},
		{
			name:        "empty file tree",
			fileTree:    []*types.FileNode{},
			wantModules: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modules := detectModules(tt.fileTree)

			if len(modules) != len(tt.wantModules) {
				t.Errorf("detectModules() returned %d modules, want %d", len(modules), len(tt.wantModules))
				return
			}

			for i, module := range modules {
				if module != tt.wantModules[i] {
					t.Errorf("Module[%d] = %v, want %v", i, module, tt.wantModules[i])
				}
			}
		})
	}
}

func TestCategorizeAABSizes(t *testing.T) {
	tests := []struct {
		name      string
		fileTree  []*types.FileNode
		wantCheck func(types.SizeBreakdown) error
	}{
		{
			name: "module structure",
			fileTree: []*types.FileNode{
				{
					Name:  "base",
					Size:  3000,
					IsDir: true,
					Children: []*types.FileNode{
						{Name: "dex", Size: 1000, IsDir: true},
						{Name: "lib", Size: 1000, IsDir: true},
						{Name: "res", Size: 1000, IsDir: true},
					},
				},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.ByCategory["Module: base"] != 3000 {
					t.Errorf("ByCategory[Module: base] = %d, want 3000", breakdown.ByCategory["Module: base"])
				}
				if breakdown.DEX != 1000 {
					t.Errorf("DEX = %d, want 1000", breakdown.DEX)
				}
				if breakdown.Libraries != 1000 {
					t.Errorf("Libraries = %d, want 1000", breakdown.Libraries)
				}
				if breakdown.Resources != 1000 {
					t.Errorf("Resources = %d, want 1000", breakdown.Resources)
				}
				return nil
			},
		},
		{
			name: "multiple modules",
			fileTree: []*types.FileNode{
				{
					Name:  "base",
					Size:  2000,
					IsDir: true,
					Children: []*types.FileNode{
						{Name: "dex", Size: 2000, IsDir: true},
					},
				},
				{
					Name:  "feature",
					Size:  1000,
					IsDir: true,
					Children: []*types.FileNode{
						{Name: "dex", Size: 1000, IsDir: true},
					},
				},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.ByCategory["Module: base"] != 2000 {
					t.Errorf("ByCategory[Module: base] = %d, want 2000", breakdown.ByCategory["Module: base"])
				}
				if breakdown.ByCategory["Module: feature"] != 1000 {
					t.Errorf("ByCategory[Module: feature] = %d, want 1000", breakdown.ByCategory["Module: feature"])
				}
				if breakdown.DEX != 3000 {
					t.Errorf("DEX = %d, want 3000", breakdown.DEX)
				}
				return nil
			},
		},
		{
			name: "DEX files in module",
			fileTree: []*types.FileNode{
				{
					Name:  "base",
					Size:  1500,
					IsDir: true,
					Children: []*types.FileNode{
						{
							Name:  "dex",
							Size:  1500,
							IsDir: true,
							Children: []*types.FileNode{
								{Name: "classes.dex", Size: 1000, IsDir: false},
								{Name: "classes2.dex", Size: 500, IsDir: false},
							},
						},
					},
				},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.DEX != 1500 {
					t.Errorf("DEX = %d, want 1500", breakdown.DEX)
				}
				if breakdown.ByCategory["DEX Files"] != 1500 {
					t.Errorf("ByCategory[DEX Files] = %d, want 1500", breakdown.ByCategory["DEX Files"])
				}
				return nil
			},
		},
		{
			name: "native libraries in module",
			fileTree: []*types.FileNode{
				{
					Name:  "base",
					Size:  2000,
					IsDir: true,
					Children: []*types.FileNode{
						{
							Name:  "lib",
							Size:  2000,
							IsDir: true,
							Children: []*types.FileNode{
								{Name: "arm64-v8a", Size: 1000, IsDir: true},
								{Name: "armeabi-v7a", Size: 1000, IsDir: true},
							},
						},
					},
				},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.Libraries != 2000 {
					t.Errorf("Libraries = %d, want 2000", breakdown.Libraries)
				}
				if breakdown.ByCategory["Native Libraries"] != 2000 {
					t.Errorf("ByCategory[Native Libraries] = %d, want 2000", breakdown.ByCategory["Native Libraries"])
				}
				return nil
			},
		},
		{
			name: "resources in module",
			fileTree: []*types.FileNode{
				{
					Name:  "base",
					Size:  3000,
					IsDir: true,
					Children: []*types.FileNode{
						{
							Name:  "res",
							Size:  3000,
							IsDir: true,
							Children: []*types.FileNode{
								{Name: "drawable", Size: 1500, IsDir: true},
								{Name: "layout", Size: 1500, IsDir: true},
							},
						},
					},
				},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.Resources != 3000 {
					t.Errorf("Resources = %d, want 3000", breakdown.Resources)
				}
				if breakdown.ByCategory["Resources"] != 3000 {
					t.Errorf("ByCategory[Resources] = %d, want 3000", breakdown.ByCategory["Resources"])
				}
				return nil
			},
		},
		{
			name: "assets in module",
			fileTree: []*types.FileNode{
				{
					Name:  "base",
					Size:  1500,
					IsDir: true,
					Children: []*types.FileNode{
						{
							Name:  "assets",
							Size:  1500,
							IsDir: true,
							Children: []*types.FileNode{
								{Name: "data.json", Size: 1500, IsDir: false},
							},
						},
					},
				},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.Assets != 1500 {
					t.Errorf("Assets = %d, want 1500", breakdown.Assets)
				}
				if breakdown.ByCategory["Assets"] != 1500 {
					t.Errorf("ByCategory[Assets] = %d, want 1500", breakdown.ByCategory["Assets"])
				}
				return nil
			},
		},
		{
			name: "bundle config and manifest",
			fileTree: []*types.FileNode{
				{Name: "BundleConfig.pb", Size: 500, IsDir: false},
				{Name: "base", Size: 1000, IsDir: true, Children: []*types.FileNode{
					{Name: "manifest", Size: 1000, IsDir: true, Children: []*types.FileNode{
						{Name: "AndroidManifest.xml", Size: 1000, IsDir: false},
					}},
				}},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				// BundleConfig.pb at root should be categorized as Resources
				if breakdown.Resources < 500 {
					t.Errorf("Resources = %d, want at least 500", breakdown.Resources)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breakdown := categorizeAABSizes(tt.fileTree)
			if err := tt.wantCheck(breakdown); err != nil {
				t.Error(err)
			}
		})
	}
}

// Helper functions to create mock AAB files

func createMockAAB(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Add BundleConfig.pb
	configWriter, err := zipWriter.Create("BundleConfig.pb")
	if err != nil {
		return err
	}
	if _, err := configWriter.Write([]byte("mock bundle config")); err != nil {
		return err
	}

	// Add base module with typical structure
	baseFiles := []struct {
		path string
		size int
	}{
		{"base/manifest/AndroidManifest.xml", 1000},
		{"base/dex/classes.dex", 1500},
		{"base/dex/classes2.dex", 800},
		{"base/lib/arm64-v8a/libnative.so", 2000},
		{"base/lib/armeabi-v7a/libnative.so", 1800},
		{"base/res/drawable/icon.png", 500},
		{"base/res/layout/main.xml", 300},
		{"base/assets/data.json", 200},
		{"base/resources.pb", 600},
	}

	for _, file := range baseFiles {
		writer, err := zipWriter.Create(file.path)
		if err != nil {
			return err
		}
		content := make([]byte, file.size)
		if _, err := writer.Write(content); err != nil {
			return err
		}
	}

	// Add a feature module
	featureFiles := []struct {
		path string
		size int
	}{
		{"feature1/manifest/AndroidManifest.xml", 500},
		{"feature1/dex/classes.dex", 800},
		{"feature1/res/drawable/feature_icon.png", 300},
	}

	for _, file := range featureFiles {
		writer, err := zipWriter.Create(file.path)
		if err != nil {
			return err
		}
		content := make([]byte, file.size)
		if _, err := writer.Write(content); err != nil {
			return err
		}
	}

	return nil
}
