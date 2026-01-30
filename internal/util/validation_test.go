package util

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/testutil"
)

func TestValidateFileArtifact(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create test files
	validFile := testutil.CreateTestFile(t, tempDir, "test.ipa", 100)
	dirPath := testutil.CreateTestDir(t, tempDir, "test.app")

	tests := []struct {
		name        string
		path        string
		expectedExt string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid file with .ipa extension",
			path:        validFile,
			expectedExt: ".ipa",
			wantErr:     false,
		},
		{
			name:        "valid file with extension without dot",
			path:        validFile,
			expectedExt: "ipa",
			wantErr:     false,
		},
		{
			name:        "wrong extension",
			path:        validFile,
			expectedExt: ".apk",
			wantErr:     true,
			errContains: "must have .apk extension",
		},
		{
			name:        "directory instead of file",
			path:        dirPath,
			expectedExt: ".app",
			wantErr:     true,
			errContains: "is a directory",
		},
		{
			name:        "nonexistent file",
			path:        tempDir + "/nonexistent.ipa",
			expectedExt: ".ipa",
			wantErr:     true,
			errContains: "failed to stat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileArtifact(tt.path, tt.expectedExt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateFileArtifact() error = %v, should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidateDirectoryArtifact(t *testing.T) {
	tempDir := testutil.CreateTempDir(t)

	// Create test files and directories
	validDir := testutil.CreateTestDir(t, tempDir, "test.app")
	filePath := testutil.CreateTestFile(t, tempDir, "test.ipa", 100)

	tests := []struct {
		name        string
		path        string
		expectedExt string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid directory with .app extension",
			path:        validDir,
			expectedExt: ".app",
			wantErr:     false,
		},
		{
			name:        "valid directory with extension without dot",
			path:        validDir,
			expectedExt: "app",
			wantErr:     false,
		},
		{
			name:        "wrong extension",
			path:        validDir,
			expectedExt: ".xcarchive",
			wantErr:     true,
			errContains: "must have .xcarchive extension",
		},
		{
			name:        "file instead of directory",
			path:        filePath,
			expectedExt: ".ipa",
			wantErr:     true,
			errContains: "is a file",
		},
		{
			name:        "nonexistent directory",
			path:        tempDir + "/nonexistent.app",
			expectedExt: ".app",
			wantErr:     true,
			errContains: "failed to stat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectoryArtifact(tt.path, tt.expectedExt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectoryArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateDirectoryArtifact() error = %v, should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findInString(s, substr)
}

func findInString(s, target string) bool {
	for i := 0; i <= len(s)-len(target); i++ {
		if s[i:i+len(target)] == target {
			return true
		}
	}
	return false
}
