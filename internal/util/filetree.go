package util

import (
	"archive/zip"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// BuildZipFileTree constructs a file tree from ZIP archive contents
func BuildZipFileTree(zipReader *zip.Reader) ([]*types.FileNode, int64) {
	// Build a map of all files first
	fileMap := make(map[string]*types.FileNode)
	var totalSize int64

	// Create nodes for all files
	for _, f := range zipReader.File {
		parts := strings.Split(f.Name, "/")
		currentPath := ""

		for i, part := range parts {
			if part == "" {
				continue
			}

			parentPath := currentPath
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = currentPath + "/" + part
			}

			// Check if node already exists
			if _, exists := fileMap[currentPath]; exists {
				continue
			}

			// Create node
			isDir := i < len(parts)-1 || f.FileInfo().IsDir()
			node := &types.FileNode{
				Path:     currentPath,
				Name:     part,
				IsDir:    isDir,
				Children: []*types.FileNode{},
			}

			if !isDir {
				node.Size = int64(f.UncompressedSize64)
				totalSize += node.Size
			}

			fileMap[currentPath] = node

			// Link to parent
			if parentPath != "" {
				if parent, exists := fileMap[parentPath]; exists {
					parent.Children = append(parent.Children, node)
				}
			}
		}
	}

	// Calculate directory sizes and collect root nodes
	var rootNodes []*types.FileNode
	for _, node := range fileMap {
		if !strings.Contains(node.Path, "/") {
			rootNodes = append(rootNodes, node)
		}
		if node.IsDir {
			node.Size = CalculateDirectorySize(node)
		}
	}

	return rootNodes, totalSize
}

// CalculateDirectorySize recursively calculates directory size
func CalculateDirectorySize(node *types.FileNode) int64 {
	if !node.IsDir {
		return node.Size
	}

	var total int64
	for _, child := range node.Children {
		if child.IsDir {
			total += CalculateDirectorySize(child)
		} else {
			total += child.Size
		}
	}
	return total
}

// FindLargestFiles returns the N largest files from the tree
func FindLargestFiles(nodes []*types.FileNode, n int) []types.FileNode {
	var files []types.FileNode

	var collectFiles func(node *types.FileNode)
	collectFiles = func(node *types.FileNode) {
		if node.IsDir {
			for _, child := range node.Children {
				collectFiles(child)
			}
		} else {
			files = append(files, *node)
		}
	}

	for _, node := range nodes {
		collectFiles(node)
	}

	// Sort by size descending
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	// Return top N
	if len(files) > n {
		files = files[:n]
	}

	return files
}
