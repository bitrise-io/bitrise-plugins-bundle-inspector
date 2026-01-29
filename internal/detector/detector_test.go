package detector

import (
	"testing"
)

func TestPathMapperToRelative(t *testing.T) {
	tests := []struct {
		name         string
		rootPath     string
		absolutePath string
		want         string
	}{
		{
			name:         "simple path",
			rootPath:     "/tmp/extracted",
			absolutePath: "/tmp/extracted/file.txt",
			want:         "file.txt",
		},
		{
			name:         "nested path",
			rootPath:     "/tmp/extracted",
			absolutePath: "/tmp/extracted/dir/subdir/file.png",
			want:         "dir/subdir/file.png",
		},
		{
			name:         "root path without trailing slash",
			rootPath:     "/tmp/extracted",
			absolutePath: "/tmp/extracted/file.txt",
			want:         "file.txt",
		},
		{
			name:         "root path with trailing slash",
			rootPath:     "/tmp/extracted/",
			absolutePath: "/tmp/extracted/file.txt",
			want:         "file.txt",
		},
		{
			name:         "path equal to root",
			rootPath:     "/tmp/extracted",
			absolutePath: "/tmp/extracted",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewPathMapper(tt.rootPath)
			got := mapper.ToRelative(tt.absolutePath)
			if got != tt.want {
				t.Errorf("ToRelative() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathMapperToRelativePaths(t *testing.T) {
	mapper := NewPathMapper("/tmp/extracted")
	input := []string{
		"/tmp/extracted/file1.txt",
		"/tmp/extracted/dir/file2.png",
		"/tmp/extracted/a/b/c/file3.jpg",
	}
	want := []string{
		"file1.txt",
		"dir/file2.png",
		"a/b/c/file3.jpg",
	}

	got := mapper.ToRelativePaths(input)

	if len(got) != len(want) {
		t.Fatalf("Expected %d paths, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Path %d: got %v, want %v", i, got[i], want[i])
		}
	}
}
