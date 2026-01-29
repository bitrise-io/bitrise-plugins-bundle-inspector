package ios

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer/ios/macho"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// ConvertToTypesBinaryInfo converts macho.BinaryInfo to types.BinaryInfo
func ConvertToTypesBinaryInfo(info *macho.BinaryInfo) *types.BinaryInfo {
	if info == nil {
		return nil
	}
	return &types.BinaryInfo{
		Architecture:     info.Architecture,
		Architectures:    info.Architectures,
		Type:             info.Type,
		CodeSize:         info.CodeSize,
		DataSize:         info.DataSize,
		LinkedLibraries:  info.LinkedLibraries,
		RPaths:           info.RPaths,
		HasDebugSymbols:  info.HasDebugSymbols,
		DebugSymbolsSize: info.DebugSymbolsSize,
	}
}

// ConvertToMachoBinaryInfo converts types.BinaryInfo to macho.BinaryInfo
func ConvertToMachoBinaryInfo(info *types.BinaryInfo) *macho.BinaryInfo {
	if info == nil {
		return nil
	}
	return &macho.BinaryInfo{
		Architecture:     info.Architecture,
		Architectures:    info.Architectures,
		Type:             info.Type,
		CodeSize:         info.CodeSize,
		DataSize:         info.DataSize,
		LinkedLibraries:  info.LinkedLibraries,
		RPaths:           info.RPaths,
		HasDebugSymbols:  info.HasDebugSymbols,
		DebugSymbolsSize: info.DebugSymbolsSize,
	}
}

// ConvertBinariesMapToMacho converts map of types.BinaryInfo to map of macho.BinaryInfo
func ConvertBinariesMapToMacho(binaries map[string]*types.BinaryInfo) map[string]*macho.BinaryInfo {
	result := make(map[string]*macho.BinaryInfo)
	for path, binInfo := range binaries {
		result[path] = ConvertToMachoBinaryInfo(binInfo)
	}
	return result
}
