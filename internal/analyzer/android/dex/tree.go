package dex

import (
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// BuildVirtualDEXTree builds a hierarchical virtual directory tree from DEX classes.
func BuildVirtualDEXTree(mergedInfo *types.MergedDEXInfo, totalDEXSize int64) *types.FileNode {
	root := &types.FileNode{
		Name:      "Dex",
		Path:      "Dex",
		IsDir:     true,
		IsVirtual: true,
		Children:  make([]*types.FileNode, 0),
		Metadata: map[string]interface{}{
			"class_count":    len(mergedInfo.Classes),
			"dex_file_count": mergedInfo.DEXFileCount,
		},
	}

	// Build tree from classes
	for _, class := range mergedInfo.Classes {
		addClassToTree(root, class)
	}

	// Calculate directory sizes (sum of children)
	calculateDirSizes(root)

	// Add _Unmapped node for shared data
	unmappedSize := totalDEXSize - mergedInfo.TotalPrivateSize
	if unmappedSize > 0 {
		unmappedNode := &types.FileNode{
			Name:      "_Unmapped",
			Path:      "Dex/_Unmapped",
			IsDir:     false,
			IsVirtual: true,
			Size:      unmappedSize,
			Metadata: map[string]interface{}{
				"description": "Shared data structures (string pools, type descriptors, proto signatures)",
				"class_type":  "unmapped_dex",
			},
		}
		root.Children = append(root.Children, unmappedNode)
		root.Size += unmappedSize
	}

	return root
}

// addClassToTree adds a class to the virtual directory tree.
func addClassToTree(root *types.FileNode, class types.DexClass) {
	// Convert package path to directory structure
	// Example: com/example/app -> ["com", "example", "app", "MainActivity.class"]
	parts := []string{}
	if class.PackageName != "" {
		parts = strings.Split(class.PackageName, "/")
	}
	parts = append(parts, class.ClassName+".class")

	// Navigate/create directory structure
	current := root
	currentPath := "Dex"

	for i, part := range parts {
		currentPath = currentPath + "/" + part
		isLeaf := (i == len(parts)-1)

		// Find or create child node
		child := findChildByName(current.Children, part)
		if child == nil {
			child = &types.FileNode{
				Name:      part,
				Path:      currentPath,
				IsDir:     !isLeaf,
				IsVirtual: true,
			}

			if isLeaf {
				// Leaf node (class file)
				child.Size = class.PrivateSize
				child.Metadata = map[string]interface{}{
					"source_dex":   class.SourceDEX,
					"method_count": class.MethodCount,
					"field_count":  class.FieldCount,
					"class_type":   "dex_class",
				}
			} else {
				// Directory node
				child.Children = make([]*types.FileNode, 0)
			}

			current.Children = append(current.Children, child)
		}

		current = child
	}
}

// findChildByName finds a child node by name.
func findChildByName(children []*types.FileNode, name string) *types.FileNode {
	for _, child := range children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// calculateDirSizes calculates directory sizes as sum of children.
func calculateDirSizes(node *types.FileNode) int64 {
	if !node.IsDir {
		return node.Size
	}

	var total int64
	for _, child := range node.Children {
		total += calculateDirSizes(child)
	}
	node.Size = total
	return total
}
