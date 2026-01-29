package macho

import (
	"debug/macho"
	"encoding/binary"
	"os"
)

// Mach-O magic numbers
const (
	MagicFat64 uint32 = 0xcafebabf
	MagicFat32 uint32 = 0xcafebabe
	Magic64    uint32 = 0xfeedfacf
	Magic32    uint32 = 0xfeedface
	Cigam64    uint32 = 0xcffaedfe
	Cigam32    uint32 = 0xcefaedfe
)

// IsMachO checks if file is a Mach-O binary by magic bytes
func IsMachO(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	var magic uint32
	if err := binary.Read(file, binary.BigEndian, &magic); err != nil {
		return false
	}

	// Check for all Mach-O magic numbers
	switch magic {
	case MagicFat64, MagicFat32, Magic64, Magic32:
		return true
	}

	// Check little-endian variants
	if magic == Cigam64 || magic == Cigam32 {
		return true
	}

	return false
}

// GetCPUTypeName converts macho.Cpu to human-readable name
func GetCPUTypeName(cpu macho.Cpu) string {
	switch cpu {
	case macho.CpuAmd64:
		return "x86_64"
	case macho.Cpu386:
		return "i386"
	case macho.CpuArm:
		return "arm"
	case macho.CpuArm64:
		return "arm64"
	case macho.CpuPpc:
		return "ppc"
	case macho.CpuPpc64:
		return "ppc64"
	default:
		return "unknown"
	}
}

// IsFatBinary checks if binary contains multiple architectures
func IsFatBinary(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	var magic uint32
	if err := binary.Read(file, binary.BigEndian, &magic); err != nil {
		return false
	}

	return magic == MagicFat64 || magic == MagicFat32
}
