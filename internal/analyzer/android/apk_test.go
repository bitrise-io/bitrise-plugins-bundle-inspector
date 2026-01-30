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

func TestAPKAnalyzer_ValidateArtifact(t *testing.T) {
	tmpDir := testutil.CreateTempDir(t)

	tests := []struct {
		name      string
		setupPath func() string
		wantErr   bool
	}{
		{
			name: "valid APK file",
			setupPath: func() string {
				return testutil.CreateTestFile(t, tmpDir, "test.apk", 100)
			},
			wantErr: false,
		},
		{
			name: "invalid extension",
			setupPath: func() string {
				return testutil.CreateTestFile(t, tmpDir, "test.txt", 100)
			},
			wantErr: true,
		},
		{
			name: "nonexistent file",
			setupPath: func() string {
				return filepath.Join(tmpDir, "nonexistent.apk")
			},
			wantErr: true,
		},
		{
			name: "directory instead of file",
			setupPath: func() string {
				dir := filepath.Join(tmpDir, "test_dir.apk")
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
				return testutil.CreateTestFile(t, tmpDir, "test_upper.APK", 100)
			},
			wantErr: false,
		},
	}

	analyzer := NewAPKAnalyzer()
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

func TestAPKAnalyzer_Analyze(t *testing.T) {
	tmpDir := testutil.CreateTempDir(t)

	// Create a mock APK file (ZIP with APK structure)
	apkPath := filepath.Join(tmpDir, "test.apk")
	if err := createMockAPK(apkPath); err != nil {
		t.Fatalf("Failed to create mock APK: %v", err)
	}

	analyzer := NewAPKAnalyzer()
	ctx := context.Background()

	report, err := analyzer.Analyze(ctx, apkPath)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Validate report structure
	if report == nil {
		t.Fatal("Expected report, got nil")
	}

	if report.ArtifactInfo.Type != types.ArtifactTypeAPK {
		t.Errorf("ArtifactType = %v, want %v", report.ArtifactInfo.Type, types.ArtifactTypeAPK)
	}

	if report.ArtifactInfo.Path != apkPath {
		t.Errorf("Path = %v, want %v", report.ArtifactInfo.Path, apkPath)
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

	// Check metadata for manifest presence
	if report.Metadata == nil {
		t.Error("Expected metadata, got nil")
	}
}

func TestAPKAnalyzer_Analyze_InvalidFile(t *testing.T) {
	tmpDir := testutil.CreateTempDir(t)

	// Create a non-ZIP file with .apk extension
	apkPath := testutil.CreateTestFile(t, tmpDir, "invalid.apk", 100)

	analyzer := NewAPKAnalyzer()
	ctx := context.Background()

	_, err := analyzer.Analyze(ctx, apkPath)
	if err == nil {
		t.Error("Expected error for invalid APK file, got nil")
	}
}

func TestParseManifest(t *testing.T) {
	tmpDir := testutil.CreateTempDir(t)

	tests := []struct {
		name            string
		includeManifest bool
		wantHasManifest bool
	}{
		{
			name:            "APK with manifest",
			includeManifest: true,
			wantHasManifest: true,
		},
		{
			name:            "APK without manifest",
			includeManifest: false,
			wantHasManifest: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apkPath := filepath.Join(tmpDir, tt.name+".apk")
			if err := createMockAPKWithOptions(apkPath, tt.includeManifest); err != nil {
				t.Fatalf("Failed to create mock APK: %v", err)
			}

			manifest, err := parseManifest(apkPath)
			if err != nil {
				t.Fatalf("parseManifest() error = %v", err)
			}

			hasManifest := manifest["has_manifest"] == true
			if hasManifest != tt.wantHasManifest {
				t.Errorf("has_manifest = %v, want %v", hasManifest, tt.wantHasManifest)
			}
		})
	}
}

func TestCategorizeAPKSizes(t *testing.T) {
	tests := []struct {
		name      string
		fileTree  []*types.FileNode
		wantCheck func(types.SizeBreakdown) error
	}{
		{
			name: "DEX files",
			fileTree: []*types.FileNode{
				{Name: "classes.dex", Size: 1000, IsDir: false},
				{Name: "classes2.dex", Size: 500, IsDir: false},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.DEX != 1500 {
					t.Errorf("DEX = %d, want 1500", breakdown.DEX)
				}
				if breakdown.ByCategory["DEX Files"] != 1500 {
					t.Errorf("ByCategory[DEX Files] = %d, want 1500", breakdown.ByCategory["DEX Files"])
				}
				if breakdown.ByExtension[".dex"] != 1500 {
					t.Errorf("ByExtension[.dex] = %d, want 1500", breakdown.ByExtension[".dex"])
				}
				return nil
			},
		},
		{
			name: "native libraries",
			fileTree: []*types.FileNode{
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
			name: "resources",
			fileTree: []*types.FileNode{
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
			name: "assets",
			fileTree: []*types.FileNode{
				{
					Name:  "assets",
					Size:  1500,
					IsDir: true,
					Children: []*types.FileNode{
						{Name: "data.json", Size: 1500, IsDir: false},
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
			name: "image resources by extension",
			fileTree: []*types.FileNode{
				{Name: "icon.png", Size: 500, IsDir: false},
				{Name: "banner.jpg", Size: 300, IsDir: false},
				{Name: "logo.webp", Size: 200, IsDir: false},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				total := int64(1000)
				if breakdown.Resources != total {
					t.Errorf("Resources = %d, want %d", breakdown.Resources, total)
				}
				if breakdown.ByExtension[".png"] != 500 {
					t.Errorf("ByExtension[.png] = %d, want 500", breakdown.ByExtension[".png"])
				}
				if breakdown.ByExtension[".jpg"] != 300 {
					t.Errorf("ByExtension[.jpg] = %d, want 300", breakdown.ByExtension[".jpg"])
				}
				if breakdown.ByExtension[".webp"] != 200 {
					t.Errorf("ByExtension[.webp] = %d, want 200", breakdown.ByExtension[".webp"])
				}
				return nil
			},
		},
		{
			name: "AndroidManifest.xml",
			fileTree: []*types.FileNode{
				{Name: "AndroidManifest.xml", Size: 1000, IsDir: false},
			},
			wantCheck: func(breakdown types.SizeBreakdown) error {
				if breakdown.Resources != 1000 {
					t.Errorf("Resources = %d, want 1000", breakdown.Resources)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breakdown := categorizeAPKSizes(tt.fileTree)
			if err := tt.wantCheck(breakdown); err != nil {
				t.Error(err)
			}
		})
	}
}

// Helper functions to create mock APK files

func createMockAPK(path string) error {
	return createMockAPKWithOptions(path, true)
}

func createMockAPKWithOptions(path string, includeManifest bool) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// Add AndroidManifest.xml if requested
	if includeManifest {
		manifestWriter, err := zipWriter.Create("AndroidManifest.xml")
		if err != nil {
			return err
		}
		// Write minimal content (binary XML in real APKs, but text is fine for testing)
		if _, err := manifestWriter.Write([]byte("<?xml version=\"1.0\" encoding=\"utf-8\"?><manifest/>")); err != nil {
			return err
		}
	}

	// Add some DEX files
	for _, name := range []string{"classes.dex", "classes2.dex"} {
		dexWriter, err := zipWriter.Create(name)
		if err != nil {
			return err
		}
		// Write some content to make the file non-empty
		content := make([]byte, 1000)
		if _, err := dexWriter.Write(content); err != nil {
			return err
		}
	}

	// Add lib directory with .so files
	for _, arch := range []string{"arm64-v8a", "armeabi-v7a"} {
		soWriter, err := zipWriter.Create("lib/" + arch + "/libnative.so")
		if err != nil {
			return err
		}
		content := make([]byte, 500)
		if _, err := soWriter.Write(content); err != nil {
			return err
		}
	}

	// Add resources
	for _, resFile := range []string{"res/drawable/icon.png", "res/layout/main.xml"} {
		resWriter, err := zipWriter.Create(resFile)
		if err != nil {
			return err
		}
		content := make([]byte, 200)
		if _, err := resWriter.Write(content); err != nil {
			return err
		}
	}

	// Add assets
	assetsWriter, err := zipWriter.Create("assets/data.json")
	if err != nil {
		return err
	}
	if _, err := assetsWriter.Write([]byte("{}")); err != nil {
		return err
	}

	return nil
}
