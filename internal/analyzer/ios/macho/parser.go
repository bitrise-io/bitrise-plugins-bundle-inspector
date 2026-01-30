package macho

import (
	"debug/macho"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ParseMachO parses a Mach-O binary and extracts metadata
func ParseMachO(path string) (*types.BinaryInfo, error) {
	file, err := macho.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Mach-O file: %w", err)
	}
	defer file.Close()

	info := &types.BinaryInfo{
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
		info.DebugSymbolsSize = estimateSymbolTableSize(file, path)
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

// estimateSymbolTableSize measures the size of strippable symbol data by running
// the strip command on a temporary copy of the binary and measuring the size difference.
//
// This is the most accurate method as it matches exactly what strip will do, correctly
// handling both executables (which can strip undefined symbols) and libraries (which
// preserve symbols needed for dynamic linking).
//
// Returns 0 if:
// - No symbol table exists
// - Strip command fails or is unavailable
// - Binary is already stripped (no savings)
func estimateSymbolTableSize(file *macho.File, path string) int64 {
	if file.Symtab == nil {
		return 0
	}

	// Measure actual strip savings
	savings, err := measureActualStripSavings(path)
	if err != nil {
		// Strip failed - binary likely already stripped or strip unavailable
		return 0
	}

	return savings
}

// measureActualStripSavings runs strip on a temporary copy and measures size difference
// This is the most accurate method as it matches exactly what strip will do
func measureActualStripSavings(binaryPath string) (int64, error) {
	// Get original size
	info, err := os.Stat(binaryPath)
	if err != nil {
		return 0, err
	}
	originalSize := info.Size()

	// Create temp copy
	tmpFile, err := os.CreateTemp("", "strip_test_*.tmp")
	if err != nil {
		return 0, err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Copy binary to temp location
	src, err := os.Open(binaryPath)
	if err != nil {
		tmpFile.Close()
		return 0, err
	}

	if _, err := io.Copy(tmpFile, src); err != nil {
		src.Close()
		tmpFile.Close()
		return 0, err
	}
	src.Close()
	tmpFile.Close()

	// Run strip (ignore warnings about code signatures)
	cmd := exec.Command("strip", "-rSTx", tmpPath)
	_ = cmd.Run() // Ignore errors - strip may warn but still work

	// Get stripped size
	strippedInfo, err := os.Stat(tmpPath)
	if err != nil {
		return 0, err
	}

	savings := originalSize - strippedInfo.Size()
	if savings < 0 {
		return 0, nil // Shouldn't happen, but protect against negative
	}

	return savings, nil
}
