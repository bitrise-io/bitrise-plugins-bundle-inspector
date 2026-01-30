package ios

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestFindMainBinary_IgnoresMetadataFiles(t *testing.T) {
	tests := []struct {
		name  string
		nodes []*types.FileNode
		want  string
	}{
		{
			name: "ignores PkgInfo, selects actual executable",
			nodes: []*types.FileNode{
				{Name: "PkgInfo", Path: "PkgInfo", Size: 8, IsDir: false},
				{Name: "Runner", Path: "Runner", Size: 1024000, IsDir: false},
				{Name: "Info.plist", Path: "Info.plist", Size: 500, IsDir: false},
			},
			want: "Runner",
		},
		{
			name: "selects largest when multiple candidates",
			nodes: []*types.FileNode{
				{Name: "Small", Path: "Small", Size: 100, IsDir: false},
				{Name: "Large", Path: "Large", Size: 100000, IsDir: false},
				{Name: "PkgInfo", Path: "PkgInfo", Size: 8, IsDir: false},
			},
			want: "Large",
		},
		{
			name: "ignores all metadata files",
			nodes: []*types.FileNode{
				{Name: "PkgInfo", Path: "PkgInfo", Size: 8, IsDir: false},
				{Name: "CodeResources", Path: "CodeResources", Size: 100, IsDir: false},
				{Name: "_CodeSignature", Path: "_CodeSignature", Size: 50, IsDir: false},
				{Name: "embedded.mobileprovision", Path: "embedded.mobileprovision", Size: 200, IsDir: false},
				{Name: "AppExecutable", Path: "AppExecutable", Size: 500000, IsDir: false},
			},
			want: "AppExecutable",
		},
		{
			name: "returns empty when only metadata files present",
			nodes: []*types.FileNode{
				{Name: "PkgInfo", Path: "PkgInfo", Size: 8, IsDir: false},
				{Name: "CodeResources", Path: "CodeResources", Size: 100, IsDir: false},
			},
			want: "",
		},
		{
			name: "returns empty when no extensionless files",
			nodes: []*types.FileNode{
				{Name: "Info.plist", Path: "Info.plist", Size: 500, IsDir: false},
				{Name: "Assets.car", Path: "Assets.car", Size: 100000, IsDir: false},
			},
			want: "",
		},
		{
			name: "handles single valid executable",
			nodes: []*types.FileNode{
				{Name: "Wikipedia", Path: "Wikipedia", Size: 5000000, IsDir: false},
			},
			want: "Wikipedia",
		},
		{
			name: "ignores directories",
			nodes: []*types.FileNode{
				{Name: "Frameworks", Path: "Frameworks", Size: 0, IsDir: true},
				{Name: "Runner", Path: "Runner", Size: 1000000, IsDir: false},
			},
			want: "Runner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findMainBinary(tt.nodes)
			if got != tt.want {
				t.Errorf("findMainBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMetadataFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "PkgInfo is metadata", filename: "PkgInfo", want: true},
		{name: "CodeResources is metadata", filename: "CodeResources", want: true},
		{name: "_CodeSignature is metadata", filename: "_CodeSignature", want: true},
		{name: "embedded.mobileprovision is metadata", filename: "embedded.mobileprovision", want: true},
		{name: "Runner is not metadata", filename: "Runner", want: false},
		{name: "Wikipedia is not metadata", filename: "Wikipedia", want: false},
		{name: "App is not metadata", filename: "App", want: false},
		{name: "Info.plist is not metadata", filename: "Info.plist", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMetadataFile(tt.filename)
			if got != tt.want {
				t.Errorf("isMetadataFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}
