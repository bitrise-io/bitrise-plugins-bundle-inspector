package analyzer

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestDetectArtifactType(t *testing.T) {
	tests := []struct {
		path         string
		expectedType types.ArtifactType
		expectError  bool
	}{
		{"app.ipa", types.ArtifactTypeIPA, false},
		{"app.IPA", types.ArtifactTypeIPA, false},
		{"app.apk", types.ArtifactTypeAPK, false},
		{"app.aab", types.ArtifactTypeAAB, false},
		{"app.app", types.ArtifactTypeApp, false},
		{"app.xcarchive", types.ArtifactTypeXCArchive, false},
		{"app.zip", "", true},
		{"app.txt", "", true},
		{"app", "", true},
	}

	for _, tt := range tests {
		result, err := DetectArtifactType(tt.path)
		if tt.expectError {
			if err == nil {
				t.Errorf("DetectArtifactType(%s) expected error, got nil", tt.path)
			}
		} else {
			if err != nil {
				t.Errorf("DetectArtifactType(%s) unexpected error: %v", tt.path, err)
			}
			if result != tt.expectedType {
				t.Errorf("DetectArtifactType(%s) = %s; want %s", tt.path, result, tt.expectedType)
			}
		}
	}
}
