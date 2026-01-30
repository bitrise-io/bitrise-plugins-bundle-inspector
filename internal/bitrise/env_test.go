package bitrise

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBitriseEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		wantIsBitrise  bool
	}{
		{
			name:          "not in Bitrise",
			envVars:       map[string]string{},
			wantIsBitrise: false,
		},
		{
			name: "in Bitrise",
			envVars: map[string]string{
				"BITRISE_BUILD_NUMBER": "123",
			},
			wantIsBitrise: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			if got := IsBitriseEnvironment(); got != tt.wantIsBitrise {
				t.Errorf("IsBitriseEnvironment() = %v, want %v", got, tt.wantIsBitrise)
			}
		})
	}
}

func TestDetectBundlePath(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	ipaPath := filepath.Join(tmpDir, "test.ipa")
	apkPath := filepath.Join(tmpDir, "test.apk")
	aabPath := filepath.Join(tmpDir, "test.aab")

	// Create test files
	for _, path := range []string{ipaPath, apkPath, aabPath} {
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name        string
		envVars     map[string]string
		wantPath    string
		wantErr     bool
	}{
		{
			name:     "no env vars set",
			envVars:  map[string]string{},
			wantPath: "",
			wantErr:  true,
		},
		{
			name: "IPA path set",
			envVars: map[string]string{
				"BITRISE_IPA_PATH": ipaPath,
			},
			wantPath: ipaPath,
			wantErr:  false,
		},
		{
			name: "APK path set",
			envVars: map[string]string{
				"BITRISE_APK_PATH": apkPath,
			},
			wantPath: apkPath,
			wantErr:  false,
		},
		{
			name: "AAB path set",
			envVars: map[string]string{
				"BITRISE_AAB_PATH": aabPath,
			},
			wantPath: aabPath,
			wantErr:  false,
		},
		{
			name: "IPA takes priority over APK",
			envVars: map[string]string{
				"BITRISE_IPA_PATH": ipaPath,
				"BITRISE_APK_PATH": apkPath,
			},
			wantPath: ipaPath,
			wantErr:  false,
		},
		{
			name: "AAB takes priority over APK",
			envVars: map[string]string{
				"BITRISE_AAB_PATH": aabPath,
				"BITRISE_APK_PATH": apkPath,
			},
			wantPath: aabPath,
			wantErr:  false,
		},
		{
			name: "IPA takes priority over all",
			envVars: map[string]string{
				"BITRISE_IPA_PATH": ipaPath,
				"BITRISE_AAB_PATH": aabPath,
				"BITRISE_APK_PATH": apkPath,
			},
			wantPath: ipaPath,
			wantErr:  false,
		},
		{
			name: "file doesn't exist",
			envVars: map[string]string{
				"BITRISE_IPA_PATH": "/nonexistent/path.ipa",
			},
			wantPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			got, err := DetectBundlePath()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectBundlePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantPath {
				t.Errorf("DetectBundlePath() = %v, want %v", got, tt.wantPath)
			}
		})
	}
}

func TestGetBuildMetadata(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    BuildMetadata
	}{
		{
			name:    "no env vars",
			envVars: map[string]string{},
			want: BuildMetadata{
				BuildNumber: "",
				CommitHash:  "",
				DeployDir:   "",
			},
		},
		{
			name: "all env vars set",
			envVars: map[string]string{
				"BITRISE_BUILD_NUMBER":  "123",
				"GIT_CLONE_COMMIT_HASH": "abc123def",
				"BITRISE_DEPLOY_DIR":    "/tmp/deploy",
			},
			want: BuildMetadata{
				BuildNumber: "123",
				CommitHash:  "abc123def",
				DeployDir:   "/tmp/deploy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			got := GetBuildMetadata()
			if got.BuildNumber != tt.want.BuildNumber {
				t.Errorf("BuildNumber = %v, want %v", got.BuildNumber, tt.want.BuildNumber)
			}
			if got.CommitHash != tt.want.CommitHash {
				t.Errorf("CommitHash = %v, want %v", got.CommitHash, tt.want.CommitHash)
			}
			if got.DeployDir != tt.want.DeployDir {
				t.Errorf("DeployDir = %v, want %v", got.DeployDir, tt.want.DeployDir)
			}
		})
	}
}

func TestWriteToDeployDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		envVars  map[string]string
		filename string
		content  []byte
		wantErr  bool
	}{
		{
			name:     "no deploy dir set",
			envVars:  map[string]string{},
			filename: "test.txt",
			content:  []byte("hello"),
			wantErr:  true,
		},
		{
			name: "write successful",
			envVars: map[string]string{
				"BITRISE_DEPLOY_DIR": tmpDir,
			},
			filename: "test.txt",
			content:  []byte("hello world"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			path, err := WriteToDeployDir(tt.filename, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteToDeployDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was written
				got, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read written file: %v", err)
					return
				}
				if string(got) != string(tt.content) {
					t.Errorf("File content = %v, want %v", string(got), string(tt.content))
				}

				// Verify path is correct
				expectedPath := filepath.Join(tmpDir, tt.filename)
				if path != expectedPath {
					t.Errorf("Path = %v, want %v", path, expectedPath)
				}
			}
		})
	}
}

func TestExportToDeployDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("test content")

	// Create source file
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	deployDir := filepath.Join(tmpDir, "deploy")

	tests := []struct {
		name       string
		envVars    map[string]string
		sourcePath string
		filename   string
		wantErr    bool
	}{
		{
			name:       "no deploy dir set",
			envVars:    map[string]string{},
			sourcePath: srcFile,
			filename:   "dest.txt",
			wantErr:    true,
		},
		{
			name: "export successful",
			envVars: map[string]string{
				"BITRISE_DEPLOY_DIR": deployDir,
			},
			sourcePath: srcFile,
			filename:   "dest.txt",
			wantErr:    false,
		},
		{
			name: "source file doesn't exist",
			envVars: map[string]string{
				"BITRISE_DEPLOY_DIR": deployDir,
			},
			sourcePath: "/nonexistent/file.txt",
			filename:   "dest.txt",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			destPath, err := ExportToDeployDir(tt.sourcePath, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportToDeployDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was copied
				got, err := os.ReadFile(destPath)
				if err != nil {
					t.Errorf("Failed to read copied file: %v", err)
					return
				}
				if string(got) != string(content) {
					t.Errorf("File content = %v, want %v", string(got), string(content))
				}

				// Verify path is correct
				expectedPath := filepath.Join(deployDir, tt.filename)
				if destPath != expectedPath {
					t.Errorf("Path = %v, want %v", destPath, expectedPath)
				}
			}
		})
	}
}
