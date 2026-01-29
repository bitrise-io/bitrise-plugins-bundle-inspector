package macho

import (
	"debug/macho"
	"encoding/binary"
	"fmt"
	"io"
	"os"
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

// estimateSymbolTableSize estimates the size of strippable symbol data
// This calculates the actual strippable symbol data size (excluding __LINKEDIT segment padding)
// by determining what percentage of symbols are strippable and applying that ratio to the
// actual symbol table size from LC_SYMTAB.
//
// Note: __LINKEDIT segments have minimum size requirements (typically 32KB) and are
// aligned to page boundaries. Small binaries can have 70-95% padding in __LINKEDIT.
// This function reports only the strippable data, not the padded segment size.
//
// The string table uses deduplication and substring sharing, so we cannot simply sum
// symbol name lengths. Instead, we read the actual strsize from LC_SYMTAB and apply
// the strippable ratio.
//
// Formula: ((nsyms × 16) + strsize) × strippable_ratio
//   - nsyms: total symbol count
//   - strsize: actual string table size (with deduplication)
//   - strippable_ratio: percentage of symbols that are debug or local
func estimateSymbolTableSize(file *macho.File, path string) int64 {
	if file.Symtab == nil {
		return 0
	}

	// Get the actual string table size from LC_SYMTAB
	// We need to read this from the raw load command since debug/macho doesn't expose it
	strsize, err := getSymtabStrsize(path)
	if err != nil || strsize == 0 {
		// Fallback: estimate conservatively
		return 0
	}

	// Count what percentage of symbols are strippable
	totalSyms := int64(len(file.Symtab.Syms))
	var strippableCount int64

	for _, sym := range file.Symtab.Syms {
		// N_STAB (0xe0) indicates debug symbol
		// Not N_EXT (0x01) means local symbol
		if (sym.Type&0xe0 != 0) || (sym.Type&0x01 == 0) {
			strippableCount++
		}
	}

	if totalSyms == 0 {
		return 0
	}

	// Calculate total symbol table size (nlist entries + string table)
	totalSymbolData := (totalSyms * 16) + strsize

	// Apply strippable ratio
	strippableRatio := float64(strippableCount) / float64(totalSyms)
	strippableSize := int64(float64(totalSymbolData) * strippableRatio)

	return strippableSize
}

// getSymtabStrsize reads the strsize field from the LC_SYMTAB load command
// This is the actual size of the string table with deduplication applied
func getSymtabStrsize(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Read Mach-O header
	var magic uint32
	if err := binary.Read(file, binary.LittleEndian, &magic); err != nil {
		return 0, err
	}

	// Check magic number and read header
	var ncmds uint32

	if magic == macho.Magic64 {
		// 64-bit binary - read full header
		file.Seek(0, 0)
		var header struct {
			Magic      uint32
			Cputype    uint32
			Cpusubtype uint32
			Filetype   uint32
			Ncmds      uint32
			Sizeofcmds uint32
			Flags      uint32
			Reserved   uint32
		}
		if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
			return 0, err
		}
		ncmds = header.Ncmds
	} else if magic == macho.Magic32 {
		// 32-bit binary
		file.Seek(0, 0)
		var header struct {
			Magic      uint32
			Cputype    uint32
			Cpusubtype uint32
			Filetype   uint32
			Ncmds      uint32
			Sizeofcmds uint32
			Flags      uint32
		}
		if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
			return 0, err
		}
		ncmds = header.Ncmds
	} else {
		return 0, fmt.Errorf("not a Mach-O file")
	}

	// Read load commands to find LC_SYMTAB
	for i := uint32(0); i < ncmds; i++ {
		var cmd, cmdsize uint32
		pos, _ := file.Seek(0, io.SeekCurrent)

		if err := binary.Read(file, binary.LittleEndian, &cmd); err != nil {
			return 0, err
		}
		if err := binary.Read(file, binary.LittleEndian, &cmdsize); err != nil {
			return 0, err
		}

		if cmd == 0x2 { // LC_SYMTAB
			// Read the LC_SYMTAB structure
			file.Seek(pos, 0)
			var symtabCmd struct {
				Cmd     uint32
				Cmdsize uint32
				Symoff  uint32
				Nsyms   uint32
				Stroff  uint32
				Strsize uint32
			}
			if err := binary.Read(file, binary.LittleEndian, &symtabCmd); err != nil {
				return 0, err
			}
			return int64(symtabCmd.Strsize), nil
		}

		// Skip to next load command
		file.Seek(pos+int64(cmdsize), 0)
	}

	return 0, fmt.Errorf("LC_SYMTAB not found")
}
