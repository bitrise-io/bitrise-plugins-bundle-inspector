package macho

import (
	"debug/macho"
	"fmt"
)

// BinaryInfo contains parsed Mach-O metadata
type BinaryInfo struct {
	Architecture     string   `json:"architecture"`
	Architectures    []string `json:"architectures"`
	Type             string   `json:"type"`
	CodeSize         int64    `json:"code_size"`
	DataSize         int64    `json:"data_size"`
	LinkedLibraries  []string `json:"linked_libraries"`
	RPaths           []string `json:"rpaths,omitempty"`
	HasDebugSymbols  bool     `json:"has_debug_symbols"`
	DebugSymbolsSize int64    `json:"debug_symbols_size,omitempty"`
}

// ParseMachO parses a Mach-O binary and extracts metadata
func ParseMachO(path string) (*BinaryInfo, error) {
	file, err := macho.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Mach-O file: %w", err)
	}
	defer file.Close()

	info := &BinaryInfo{
		Architecture:    GetCPUTypeName(file.Cpu),
		Architectures:   []string{GetCPUTypeName(file.Cpu)},
		Type:            getBinaryType(file.Type),
		LinkedLibraries: make([]string, 0),
		RPaths:          make([]string, 0),
	}

	// Extract segment sizes
	codeSize, dataSize := getSegmentSizes(file)
	info.CodeSize = codeSize
	info.DataSize = dataSize

	// Extract linked libraries and rpaths
	for _, load := range file.Loads {
		switch cmd := load.(type) {
		case *macho.Dylib:
			info.LinkedLibraries = append(info.LinkedLibraries, cmd.Name)
		case *macho.Rpath:
			info.RPaths = append(info.RPaths, cmd.Path)
		}
	}

	// Check for debug symbols
	info.HasDebugSymbols = hasDebugSymbols(file)

	// Estimate debug symbol size if present
	if info.HasDebugSymbols {
		info.DebugSymbolsSize = estimateSymbolTableSize(file)
	}

	return info, nil
}

// GetArchitectures returns all architectures in binary (handles fat binaries)
func GetArchitectures(path string) ([]string, error) {
	// Try to open as fat binary first
	fatFile, err := macho.OpenFat(path)
	if err == nil {
		defer fatFile.Close()
		archs := make([]string, len(fatFile.Arches))
		for i, arch := range fatFile.Arches {
			archs[i] = GetCPUTypeName(arch.Cpu)
		}
		return archs, nil
	}

	// Not a fat binary, try regular Mach-O
	file, err := macho.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open as Mach-O or fat binary: %w", err)
	}
	defer file.Close()

	return []string{GetCPUTypeName(file.Cpu)}, nil
}

// GetLinkedLibraries extracts LC_LOAD_DYLIB load commands
func GetLinkedLibraries(path string) ([]string, error) {
	file, err := macho.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Mach-O file: %w", err)
	}
	defer file.Close()

	var libraries []string
	for _, load := range file.Loads {
		if dylib, ok := load.(*macho.Dylib); ok {
			libraries = append(libraries, dylib.Name)
		}
	}

	return libraries, nil
}

// HasDebugSymbols checks for DWARF debug info in binary
func HasDebugSymbols(path string) (bool, error) {
	file, err := macho.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open Mach-O file: %w", err)
	}
	defer file.Close()

	return hasDebugSymbols(file), nil
}

// GetSegmentSizes calculates __TEXT and __DATA segment sizes
func GetSegmentSizes(path string) (codeSize, dataSize int64, err error) {
	file, err := macho.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open Mach-O file: %w", err)
	}
	defer file.Close()

	codeSize, dataSize = getSegmentSizes(file)
	return codeSize, dataSize, nil
}

// Helper functions

func getBinaryType(fileType macho.Type) string {
	switch fileType {
	case macho.TypeExec:
		return "executable"
	case macho.TypeDylib:
		return "dylib"
	case macho.TypeBundle:
		return "bundle"
	case macho.TypeObj:
		return "object"
	default:
		return fmt.Sprintf("unknown(%d)", fileType)
	}
}

func getSegmentSizes(file *macho.File) (codeSize, dataSize int64) {
	for _, load := range file.Loads {
		if seg, ok := load.(*macho.Segment); ok {
			switch seg.Name {
			case "__TEXT":
				codeSize += int64(seg.Filesz)
			case "__DATA", "__DATA_CONST", "__DATA_DIRTY":
				dataSize += int64(seg.Filesz)
			}
		}
	}
	return codeSize, dataSize
}

func hasDebugSymbols(file *macho.File) bool {
	// Level 1: Check for DWARF debug information
	_, err := file.DWARF()
	if err == nil {
		return true
	}

	// Level 2: Check for DWARF segments manually
	if file.Sections != nil {
		debugSections := []string{
			"__debug_info",
			"__debug_line",
			"__debug_abbrev",
			"__debug_str",
		}

		for _, section := range file.Sections {
			// Check if in __DWARF segment
			if section.Seg == "__DWARF" {
				return true
			}
			// Check for debug section names
			for _, debugSection := range debugSections {
				if section.Name == debugSection {
					return true
				}
			}
		}
	}

	// Level 3: Check for symbol table entries
	// Symbol tables in __LINKEDIT segment contain debug/local symbols that can be stripped
	if file.Symtab != nil {
		for _, sym := range file.Symtab.Syms {
			// Check symbol type flags
			// N_STAB (0xe0) indicates debug symbol
			if sym.Type&0xe0 != 0 {
				return true
			}
			// Local symbols (not external) are also strippable
			// N_EXT (0x01) = external symbol (exported)
			if sym.Type&0x01 == 0 {
				return true
			}
		}
	}

	return false
}

// estimateSymbolTableSize estimates the size of strippable symbol data
func estimateSymbolTableSize(file *macho.File) int64 {
	if file.Symtab == nil {
		return 0
	}

	var totalSize int64

	// Calculate size for each strippable symbol
	// Each symbol consists of:
	// - nlist_64 entry: 16 bytes (on 64-bit architectures)
	// - string table entry: symbol name + null terminator
	for _, sym := range file.Symtab.Syms {
		isStrippable := false

		// N_STAB (0xe0) indicates debug symbol
		if sym.Type&0xe0 != 0 {
			isStrippable = true
		} else if sym.Type&0x01 == 0 { // Not external (N_EXT) - local symbol
			isStrippable = true
		}

		if isStrippable {
			// nlist_64 entry size (16 bytes for 64-bit)
			totalSize += 16

			// String table entry: symbol name length + null terminator
			nameSize := int64(len(sym.Name))
			if nameSize > 0 {
				totalSize += nameSize + 1 // +1 for null terminator
			}
		}
	}

	// Apply correction factor: not all string table space is freed
	// String tables often have deduplication and shared strings
	// Empirically, actual savings are about 75-80% of calculated size
	// This matches real-world strip -x behavior
	totalSize = (totalSize * 3) / 4 // 75% factor

	return totalSize
}
