package util

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestBuildZipFileTree(t *testing.T) {
	// Create a test ZIP archive in memory
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// Add test files
	files := []struct {
		name string
		size uint64
	}{
		{"README.md", 100},
		{"src/main.go", 500},
		{"src/util.go", 300},
		{"assets/icon.png", 1024},
		{"assets/images/logo.png", 2048},
	}

	for _, f := range files {
		header := &zip.FileHeader{
			Name:               f.name,
			UncompressedSize64: f.size,
		}
		fw, err := w.CreateHeader(header)
		if err != nil {
			t.Fatalf("Failed to create zip entry: %v", err)
		}
		// Write dummy data of the specified size
		data := make([]byte, f.size)
		if _, err := fw.Write(data); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	// Read the ZIP
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Failed to create zip reader: %v", err)
	}

	// Build file tree
	tree, totalSize := BuildZipFileTree(r)

	// Verify total size
	expectedTotal := int64(100 + 500 + 300 + 1024 + 2048)
	if totalSize != expectedTotal {
		t.Errorf("Expected total size %d, got %d", expectedTotal, totalSize)
	}

	// Verify root nodes
	if len(tree) != 3 { // README.md, src, assets
		t.Errorf("Expected 3 root nodes, got %d", len(tree))
	}

	// Find and verify the src directory
	var srcNode *types.FileNode
	for _, node := range tree {
		if node.Name == "src" {
			srcNode = node
			break
		}
	}

	if srcNode == nil {
		t.Fatal("src directory not found in tree")
	}

	if !srcNode.IsDir {
		t.Error("src should be a directory")
	}

	if len(srcNode.Children) != 2 {
		t.Errorf("Expected 2 children in src, got %d", len(srcNode.Children))
	}

	// Verify directory size calculation
	expectedSrcSize := int64(500 + 300)
	if srcNode.Size != expectedSrcSize {
		t.Errorf("Expected src size %d, got %d", expectedSrcSize, srcNode.Size)
	}
}

func TestCalculateDirectorySize(t *testing.T) {
	tests := []struct {
		name     string
		node     *types.FileNode
		expected int64
	}{
		{
			name: "file node",
			node: &types.FileNode{
				Name:  "file.txt",
				Size:  100,
				IsDir: false,
			},
			expected: 100,
		},
		{
			name: "directory with files",
			node: &types.FileNode{
				Name:  "dir",
				IsDir: true,
				Children: []*types.FileNode{
					{Name: "file1.txt", Size: 100, IsDir: false},
					{Name: "file2.txt", Size: 200, IsDir: false},
				},
			},
			expected: 300,
		},
		{
			name: "nested directories",
			node: &types.FileNode{
				Name:  "parent",
				IsDir: true,
				Children: []*types.FileNode{
					{Name: "file1.txt", Size: 100, IsDir: false},
					{
						Name:  "child",
						IsDir: true,
						Children: []*types.FileNode{
							{Name: "file2.txt", Size: 200, IsDir: false},
							{Name: "file3.txt", Size: 300, IsDir: false},
						},
					},
				},
			},
			expected: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set sizes for directories (simulate after building tree)
			if tt.node.IsDir {
				tt.node.Size = CalculateDirectorySize(tt.node)
			}

			if tt.node.Size != tt.expected {
				t.Errorf("Expected size %d, got %d", tt.expected, tt.node.Size)
			}
		})
	}
}

func TestFindLargestFiles(t *testing.T) {
	tree := []*types.FileNode{
		{
			Name:  "dir1",
			IsDir: true,
			Children: []*types.FileNode{
				{Name: "small.txt", Size: 100, IsDir: false},
				{Name: "medium.txt", Size: 500, IsDir: false},
			},
		},
		{
			Name:  "dir2",
			IsDir: true,
			Children: []*types.FileNode{
				{Name: "large.txt", Size: 1000, IsDir: false},
				{
					Name:  "subdir",
					IsDir: true,
					Children: []*types.FileNode{
						{Name: "huge.txt", Size: 5000, IsDir: false},
					},
				},
			},
		},
		{Name: "root.txt", Size: 750, IsDir: false},
	}

	tests := []struct {
		name           string
		n              int
		expectedCount  int
		expectedLargest string
		expectedSize   int64
	}{
		{
			name:           "top 3 files",
			n:              3,
			expectedCount:  3,
			expectedLargest: "huge.txt",
			expectedSize:   5000,
		},
		{
			name:           "top 10 (more than available)",
			n:              10,
			expectedCount:  5,
			expectedLargest: "huge.txt",
			expectedSize:   5000,
		},
		{
			name:           "top 1",
			n:              1,
			expectedCount:  1,
			expectedLargest: "huge.txt",
			expectedSize:   5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			largest := FindLargestFiles(tree, tt.n)

			if len(largest) != tt.expectedCount {
				t.Errorf("Expected %d files, got %d", tt.expectedCount, len(largest))
			}

			if len(largest) > 0 {
				if largest[0].Name != tt.expectedLargest {
					t.Errorf("Expected largest file to be %s, got %s", tt.expectedLargest, largest[0].Name)
				}

				if largest[0].Size != tt.expectedSize {
					t.Errorf("Expected largest file size %d, got %d", tt.expectedSize, largest[0].Size)
				}

				// Verify files are sorted by size descending
				for i := 1; i < len(largest); i++ {
					if largest[i].Size > largest[i-1].Size {
						t.Error("Files are not sorted by size descending")
						break
					}
				}
			}
		})
	}
}
