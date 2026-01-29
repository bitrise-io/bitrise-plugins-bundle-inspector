package ios

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestConvertToTypesBinaryInfo(t *testing.T) {
	tests := []struct {
		name  string
		input *macho.BinaryInfo
		want  *types.BinaryInfo
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "full conversion",
			input: &macho.BinaryInfo{
				Architecture:     "arm64",
				Architectures:    []string{"arm64", "x86_64"},
				Type:             "executable",
				CodeSize:         1024,
				DataSize:         512,
				LinkedLibraries:  []string{"Foundation", "UIKit"},
				RPaths:           []string{"@executable_path/Frameworks"},
				HasDebugSymbols:  true,
				DebugSymbolsSize: 2048,
			},
			want: &types.BinaryInfo{
				Architecture:     "arm64",
				Architectures:    []string{"arm64", "x86_64"},
				Type:             "executable",
				CodeSize:         1024,
				DataSize:         512,
				LinkedLibraries:  []string{"Foundation", "UIKit"},
				RPaths:           []string{"@executable_path/Frameworks"},
				HasDebugSymbols:  true,
				DebugSymbolsSize: 2048,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToTypesBinaryInfo(tt.input)
			if got == nil && tt.want == nil {
				return
			}
			if got == nil || tt.want == nil {
				t.Errorf("ConvertToTypesBinaryInfo() = %v, want %v", got, tt.want)
				return
			}
			if got.Architecture != tt.want.Architecture {
				t.Errorf("Architecture = %v, want %v", got.Architecture, tt.want.Architecture)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.CodeSize != tt.want.CodeSize {
				t.Errorf("CodeSize = %v, want %v", got.CodeSize, tt.want.CodeSize)
			}
			if got.HasDebugSymbols != tt.want.HasDebugSymbols {
				t.Errorf("HasDebugSymbols = %v, want %v", got.HasDebugSymbols, tt.want.HasDebugSymbols)
			}
		})
	}
}

func TestConvertToMachoBinaryInfo(t *testing.T) {
	tests := []struct {
		name  string
		input *types.BinaryInfo
		want  *macho.BinaryInfo
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "full conversion",
			input: &types.BinaryInfo{
				Architecture:     "arm64",
				Architectures:    []string{"arm64", "x86_64"},
				Type:             "dylib",
				CodeSize:         2048,
				DataSize:         1024,
				LinkedLibraries:  []string{"Foundation"},
				RPaths:           []string{"@loader_path/Frameworks"},
				HasDebugSymbols:  false,
				DebugSymbolsSize: 0,
			},
			want: &macho.BinaryInfo{
				Architecture:     "arm64",
				Architectures:    []string{"arm64", "x86_64"},
				Type:             "dylib",
				CodeSize:         2048,
				DataSize:         1024,
				LinkedLibraries:  []string{"Foundation"},
				RPaths:           []string{"@loader_path/Frameworks"},
				HasDebugSymbols:  false,
				DebugSymbolsSize: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToMachoBinaryInfo(tt.input)
			if got == nil && tt.want == nil {
				return
			}
			if got == nil || tt.want == nil {
				t.Errorf("ConvertToMachoBinaryInfo() = %v, want %v", got, tt.want)
				return
			}
			if got.Architecture != tt.want.Architecture {
				t.Errorf("Architecture = %v, want %v", got.Architecture, tt.want.Architecture)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.CodeSize != tt.want.CodeSize {
				t.Errorf("CodeSize = %v, want %v", got.CodeSize, tt.want.CodeSize)
			}
		})
	}
}

func TestConvertBinariesMapToMacho(t *testing.T) {
	input := map[string]*types.BinaryInfo{
		"App": {
			Architecture: "arm64",
			Type:         "executable",
			CodeSize:     1024,
		},
		"Framework.framework/Framework": {
			Architecture: "x86_64",
			Type:         "dylib",
			CodeSize:     2048,
		},
	}

	result := ConvertBinariesMapToMacho(input)

	if len(result) != len(input) {
		t.Errorf("Expected %d entries, got %d", len(input), len(result))
	}

	for path, binInfo := range input {
		machoInfo, exists := result[path]
		if !exists {
			t.Errorf("Missing entry for path %s", path)
			continue
		}
		if machoInfo.Architecture != binInfo.Architecture {
			t.Errorf("Path %s: Architecture = %v, want %v", path, machoInfo.Architecture, binInfo.Architecture)
		}
		if machoInfo.Type != binInfo.Type {
			t.Errorf("Path %s: Type = %v, want %v", path, machoInfo.Type, binInfo.Type)
		}
		if machoInfo.CodeSize != binInfo.CodeSize {
			t.Errorf("Path %s: CodeSize = %v, want %v", path, machoInfo.CodeSize, binInfo.CodeSize)
		}
	}
}
