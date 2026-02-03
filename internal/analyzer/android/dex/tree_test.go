package dex

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestBuildVirtualDEXTree(t *testing.T) {
	mergedInfo := &types.MergedDEXInfo{
		Classes: []types.DexClass{
			{
				ClassName:   "MainActivity",
				PackageName: "com/example/app",
				PrivateSize: 1000,
				MethodCount: 5,
				FieldCount:  3,
				SourceDEX:   "classes.dex",
			},
			{
				ClassName:   "Utils",
				PackageName: "com/example/app",
				PrivateSize: 500,
				MethodCount: 3,
				FieldCount:  1,
				SourceDEX:   "classes.dex",
			},
			{
				ClassName:   "GameEngine",
				PackageName: "com/example/game",
				PrivateSize: 2000,
				MethodCount: 10,
				FieldCount:  5,
				SourceDEX:   "classes2.dex",
			},
		},
		TotalPrivateSize: 3500,
		TotalFileSize:    5000,
		DEXFileCount:     2,
	}

	totalDEXSize := int64(5000)

	tree := BuildVirtualDEXTree(mergedInfo, totalDEXSize)

	// Verify root node
	if tree.Name != "Dex" {
		t.Errorf("Root name = %q, want %q", tree.Name, "Dex")
	}
	if !tree.IsVirtual {
		t.Error("Root should be virtual")
	}
	if !tree.IsDir {
		t.Error("Root should be a directory")
	}

	// Verify total size includes unmapped
	if tree.Size != totalDEXSize {
		t.Errorf("Root size = %d, want %d", tree.Size, totalDEXSize)
	}

	// Verify _Unmapped node exists
	var unmappedNode *types.FileNode
	for _, child := range tree.Children {
		if child.Name == "_Unmapped" {
			unmappedNode = child
			break
		}
	}
	if unmappedNode == nil {
		t.Fatal("_Unmapped node not found")
	}

	expectedUnmapped := totalDEXSize - mergedInfo.TotalPrivateSize
	if unmappedNode.Size != expectedUnmapped {
		t.Errorf("Unmapped size = %d, want %d", unmappedNode.Size, expectedUnmapped)
	}

	// Verify package structure
	var comNode *types.FileNode
	for _, child := range tree.Children {
		if child.Name == "com" {
			comNode = child
			break
		}
	}
	if comNode == nil {
		t.Fatal("com package not found")
	}
	if !comNode.IsDir {
		t.Error("com should be a directory")
	}

	// Verify class metadata
	var mainActivityNode *types.FileNode
	var findClass func([]*types.FileNode, string) *types.FileNode
	findClass = func(nodes []*types.FileNode, name string) *types.FileNode {
		for _, node := range nodes {
			if node.Name == name {
				return node
			}
			if node.IsDir && node.Children != nil {
				if found := findClass(node.Children, name); found != nil {
					return found
				}
			}
		}
		return nil
	}

	mainActivityNode = findClass(tree.Children, "MainActivity.class")
	if mainActivityNode == nil {
		t.Fatal("MainActivity.class not found")
	}

	if mainActivityNode.Size != 1000 {
		t.Errorf("MainActivity size = %d, want %d", mainActivityNode.Size, 1000)
	}

	metadata := mainActivityNode.Metadata
	if metadata == nil {
		t.Fatal("MainActivity metadata is nil")
	}

	if metadata["class_type"] != "dex_class" {
		t.Errorf("class_type = %v, want %q", metadata["class_type"], "dex_class")
	}
	if metadata["method_count"] != 5 {
		t.Errorf("method_count = %v, want %d", metadata["method_count"], 5)
	}
	if metadata["field_count"] != 3 {
		t.Errorf("field_count = %v, want %d", metadata["field_count"], 3)
	}
	if metadata["source_dex"] != "classes.dex" {
		t.Errorf("source_dex = %v, want %q", metadata["source_dex"], "classes.dex")
	}
}

func TestCalculateDirSizes(t *testing.T) {
	// Create a simple tree structure
	root := &types.FileNode{
		Name:  "root",
		IsDir: true,
		Children: []*types.FileNode{
			{
				Name:  "dir1",
				IsDir: true,
				Children: []*types.FileNode{
					{Name: "file1.txt", Size: 100},
					{Name: "file2.txt", Size: 200},
				},
			},
			{
				Name:  "dir2",
				IsDir: true,
				Children: []*types.FileNode{
					{Name: "file3.txt", Size: 300},
				},
			},
			{Name: "file4.txt", Size: 400},
		},
	}

	totalSize := calculateDirSizes(root)

	expectedTotal := int64(1000)
	if totalSize != expectedTotal {
		t.Errorf("Total size = %d, want %d", totalSize, expectedTotal)
	}

	dir1 := root.Children[0]
	if dir1.Size != 300 {
		t.Errorf("dir1 size = %d, want %d", dir1.Size, 300)
	}

	dir2 := root.Children[1]
	if dir2.Size != 300 {
		t.Errorf("dir2 size = %d, want %d", dir2.Size, 300)
	}
}

func TestFindChildByName(t *testing.T) {
	children := []*types.FileNode{
		{Name: "file1.txt"},
		{Name: "file2.txt"},
		{Name: "dir1", IsDir: true},
	}

	tests := []struct {
		name     string
		target   string
		wantName string
		wantNil  bool
	}{
		{"find first", "file1.txt", "file1.txt", false},
		{"find middle", "file2.txt", "file2.txt", false},
		{"find last", "dir1", "dir1", false},
		{"not found", "missing.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findChildByName(children, tt.target)
			if tt.wantNil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Fatalf("Expected node, got nil")
				}
				if result.Name != tt.wantName {
					t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
				}
			}
		})
	}
}

func TestBuildVirtualDEXTree_NoUnmapped(t *testing.T) {
	// Test case where private size equals total size (no unmapped data)
	mergedInfo := &types.MergedDEXInfo{
		Classes: []types.DexClass{
			{
				ClassName:   "SimpleClass",
				PackageName: "com/example",
				PrivateSize: 1000,
				MethodCount: 1,
				FieldCount:  1,
				SourceDEX:   "classes.dex",
			},
		},
		TotalPrivateSize: 1000,
		TotalFileSize:    1000,
		DEXFileCount:     1,
	}

	totalDEXSize := int64(1000)

	tree := BuildVirtualDEXTree(mergedInfo, totalDEXSize)

	// Verify no _Unmapped node when sizes match
	hasUnmapped := false
	for _, child := range tree.Children {
		if child.Name == "_Unmapped" {
			hasUnmapped = true
			break
		}
	}

	if hasUnmapped {
		t.Error("Should not have _Unmapped node when private size equals total size")
	}

	if tree.Size != totalDEXSize {
		t.Errorf("Root size = %d, want %d", tree.Size, totalDEXSize)
	}
}
