// Package dex provides DEX file parsing and analysis capabilities.
package dex

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
	"github.com/csnewman/dextk"
)

// ParseDEXFile parses a single DEX file and extracts class information.
func ParseDEXFile(path string) (*types.DexInfo, error) {
	// Open DEX file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open DEX file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat DEX file: %w", err)
	}
	fileSize := stat.Size()

	// Parse DEX file
	reader, err := dextk.Read(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DEX file: %w", err)
	}

	info := &types.DexInfo{
		SourceFile:    path,
		Classes:       make([]types.DexClass, 0),
		TotalFileSize: fileSize,
		Metadata:      make(map[string]interface{}),
	}

	// Extract class information using ClassIter
	classIter := reader.ClassIter()
	for classIter.HasNext() {
		classNode, err := classIter.Next()
		if err != nil {
			// Skip classes that fail to parse
			continue
		}

		classInfo := extractClassInfo(&classNode, path)
		info.Classes = append(info.Classes, classInfo)
		info.TotalPrivateSize += classInfo.PrivateSize
	}

	// Detect obfuscation (if many single-letter class names)
	info.IsObfuscated = detectObfuscation(info.Classes)

	// Add metadata
	info.Metadata["class_count"] = len(info.Classes)
	info.Metadata["string_pool_size"] = reader.StringIDCount
	info.Metadata["type_count"] = reader.TypeIDCount
	info.Metadata["dex_version"] = reader.Version

	return info, nil
}

// extractClassInfo extracts information about a single class from DEX.
func extractClassInfo(classNode *dextk.ClassNode, sourceDEX string) types.DexClass {
	className, packageName := parseClassName(classNode.Name.String())

	classInfo := types.DexClass{
		ClassName:   className,
		PackageName: packageName,
		SourceDEX:   sourceDEX,
		MethodCount: len(classNode.DirectMethods) + len(classNode.VirtualMethods),
		FieldCount:  len(classNode.StaticFields) + len(classNode.InstanceFields),
		PrivateSize: calculatePrivateSize(classNode),
		Metadata:    make(map[string]interface{}),
	}

	return classInfo
}

// calculatePrivateSize calculates the private size of a class.
// Private size includes only data structures 100% attributable to this class:
// - class_def entry (fixed size)
// - class_data_item (fields and methods metadata)
// - code_item structures (method bytecode)
// - class-specific annotations
//
// It does NOT include shared structures like string pools, type descriptors, or proto signatures.
func calculatePrivateSize(classNode *dextk.ClassNode) int64 {
	var size int64

	// 1. class_def entry: 32 bytes (fixed in DEX format)
	size += 32

	// 2. class_data_item: approximate size based on field/method counts
	// Each field: ~8 bytes (field_idx_diff, access_flags)
	// Each method: ~12 bytes (method_idx_diff, access_flags, code_off)
	fieldCount := len(classNode.StaticFields) + len(classNode.InstanceFields)
	methodCount := len(classNode.DirectMethods) + len(classNode.VirtualMethods)
	size += int64(fieldCount * 8)
	size += int64(methodCount * 12)

	// 3. Method bytecode: estimate based on method count
	// Each method with code typically has:
	// - code_item header: 16 bytes
	// - Instructions: average ~100 bytes per method
	// This is a conservative estimate
	methodsWithCode := 0
	for _, method := range classNode.DirectMethods {
		if method.CodeOff != 0 {
			methodsWithCode++
		}
	}
	for _, method := range classNode.VirtualMethods {
		if method.CodeOff != 0 {
			methodsWithCode++
		}
	}
	size += int64(methodsWithCode) * 116 // 16 (header) + 100 (avg bytecode)

	// 4. Annotations: estimate fixed size if class has methods/fields
	// (likely to have some annotations)
	if len(classNode.DirectMethods) > 0 || len(classNode.VirtualMethods) > 0 {
		size += 50 // Conservative annotation estimate
	}

	return size
}

// parseClassName splits a DEX class descriptor into package and class name.
// Example: "Lcom/example/app/MainActivity;" -> ("MainActivity", "com/example/app")
func parseClassName(descriptor string) (className, packageName string) {
	// Remove leading 'L' and trailing ';'
	name := strings.TrimPrefix(descriptor, "L")
	name = strings.TrimSuffix(name, ";")

	// Split on last '/'
	lastSlash := strings.LastIndex(name, "/")
	if lastSlash == -1 {
		// No package (e.g., "La;" -> "a", "")
		return name, ""
	}

	className = name[lastSlash+1:]
	packageName = name[:lastSlash]

	return className, packageName
}

// detectObfuscation checks if the DEX file appears to be obfuscated.
// Heuristic: if >50% of classes have single-letter names, likely obfuscated.
func detectObfuscation(classes []types.DexClass) bool {
	if len(classes) == 0 {
		return false
	}

	singleLetterCount := 0
	for _, class := range classes {
		if len(class.ClassName) == 1 {
			singleLetterCount++
		}
	}

	threshold := float64(len(classes)) * 0.5
	return float64(singleLetterCount) > threshold
}
