// +build integration

package main

import (
	"context"
	"os"
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer"
)

// TestRealArtifacts runs analysis on real-world artifacts
func TestRealArtifacts(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		artifactType string
	}{
		{
			name:         "iOS IPA - Lightyear",
			path:         "test-artifacts/ios/lightyear.ipa",
			artifactType: "ipa",
		},
		{
			name:         "iOS App Bundle - Wikipedia",
			path:         "test-artifacts/ios/Wikipedia.app",
			artifactType: "app",
		},
		{
			name:         "Android APK - 2048",
			path:         "test-artifacts/android/2048-game-2048.apk",
			artifactType: "apk",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if artifact doesn't exist
			if _, err := os.Stat(tc.path); os.IsNotExist(err) {
				t.Skipf("Test artifact not found: %s", tc.path)
				return
			}

			// Create analyzer
			a, err := analyzer.NewAnalyzer(tc.path)
			if err != nil {
				t.Fatalf("Failed to create analyzer: %v", err)
			}

			// Run analysis
			ctx := context.Background()
			report, err := a.Analyze(ctx, tc.path)
			if err != nil {
				t.Fatalf("Analysis failed: %v", err)
			}

			// Validate report
			if report == nil {
				t.Fatal("Report is nil")
			}
			if report.ArtifactInfo.Size == 0 {
				t.Error("Artifact size should not be 0")
			}
			if len(report.FileTree) == 0 {
				t.Error("File tree should not be empty")
			}

			// Log summary
			t.Logf("✓ Analyzed %s", tc.name)
			t.Logf("  Size: %d bytes", report.ArtifactInfo.Size)
			t.Logf("  Uncompressed: %d bytes", report.ArtifactInfo.UncompressedSize)
			t.Logf("  Files: %d", len(report.FileTree))
		})
	}
}
