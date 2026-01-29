# iOS Advanced Analysis

This document describes the advanced iOS analysis capabilities implemented in Phase 4 of the bundle-inspector project.

## Overview

The iOS advanced analysis provides deep inspection of iOS applications beyond basic file enumeration. It includes:

1. **Mach-O Binary Parsing** - Architecture detection and binary metadata extraction
2. **Framework Dependency Analysis** - Framework discovery and dependency graph construction
3. **Assets.car Parsing** - Asset catalog inspection and optimization
4. **LZFSE Compression Support** - Modern IPA decompression

## Features

### 1. Mach-O Binary Parsing

The bundle-inspector automatically detects and parses Mach-O binaries (executables, dylibs, frameworks) in iOS applications.

#### Detected Information

- **Architecture**: arm64, x86_64, i386, etc.
- **Binary Type**: executable, dylib, bundle, object file
- **Code/Data Sizes**: __TEXT and __DATA segment sizes
- **Linked Libraries**: All LC_LOAD_DYLIB dependencies
- **RPaths**: @rpath, @executable_path, @loader_path entries
- **Debug Symbols**: DWARF debug information detection

#### Example Output

```json
{
  "binaries": {
    "Wikipedia": {
      "architecture": "arm64",
      "architectures": ["arm64"],
      "type": "executable",
      "code_size": 16384,
      "data_size": 16384,
      "linked_libraries": [
        "@rpath/Wikipedia.debug.dylib",
        "/usr/lib/libSystem.B.dylib"
      ],
      "rpaths": [
        "@executable_path",
        "@executable_path/Frameworks"
      ],
      "has_debug_symbols": false
    }
  }
}
```

#### Architecture Detection

The parser uses Go's `debug/macho` package to:
- Read Mach-O headers and load commands
- Detect fat/universal binaries (multiple architectures)
- Extract CPU type and convert to human-readable names
- Handle both big-endian and little-endian formats

#### Magic Byte Detection

Before parsing, files are checked for Mach-O magic bytes:
- `0xfeedface` / `0xfeedfacf` - Mach-O 32/64-bit
- `0xcafebabe` / `0xcafebabf` - Fat binary 32/64-bit

This ensures efficient detection without attempting to parse non-Mach-O files.

### 2. Framework Dependency Analysis

The bundle-inspector discovers frameworks and builds a complete dependency graph.

#### Framework Discovery

Automatically finds all `.framework` directories in the app bundle:
- Scans `Frameworks/` directory
- Extracts framework binary and Info.plist metadata
- Reads `CFBundleShortVersionString` for version information
- Calculates total framework size

#### Dependency Graph

Builds a complete dependency graph showing framework relationships:

```json
{
  "dependency_graph": {
    "Wikipedia": ["@rpath/Wikipedia.debug.dylib"],
    "Wikipedia.debug.dylib": ["Frameworks/WMF.framework/WMF"],
    "Frameworks/WMF.framework/WMF": [],
    "PlugIns/ContinueReadingWidget.appex/ContinueReadingWidget": [
      "@rpath/ContinueReadingWidget.debug.dylib"
    ],
    "PlugIns/ContinueReadingWidget.appex/ContinueReadingWidget.debug.dylib": [
      "Frameworks/WMF.framework/WMF"
    ]
  }
}
```

#### @rpath Resolution

The analyzer resolves special path references:
- `@rpath/Framework.framework/Framework` → Resolves using LC_RPATH entries
- `@executable_path/../Frameworks/` → Relative to main binary
- `@loader_path/` → Relative to loading binary
- System libraries (`/System/Library/`, `/usr/lib/`) → Filtered from graph

#### Unused Framework Detection

Identifies frameworks not linked by any binary in the app:
- Builds transitive dependency closure
- Detects frameworks not reachable from main binary
- Generates optimization recommendations

Example optimization:

```json
{
  "category": "frameworks",
  "severity": "medium",
  "title": "Unused framework: Unused.framework",
  "description": "Framework is not linked by main binary or other frameworks",
  "action": "Remove framework to reduce app size",
  "impact": 15728640
}
```

### 3. Assets.car Parsing

The bundle-inspector parses Assets.car files to extract asset metadata.

#### Asset Catalog Information

Extracts comprehensive asset information:

```json
{
  "asset_catalogs": [
    {
      "path": "Assets.car",
      "total_size": 2955368,
      "asset_count": 331,
      "by_type": {
        "png": 89280,
        "pdf": 855368,
        "data": 8928,
        "unknown": 2001792
      },
      "by_scale": {
        "1x": 500000,
        "2x": 1800000,
        "3x": 655368
      },
      "largest_assets": [
        {
          "name": "AppIcon",
          "type": "png",
          "scale": "3x",
          "size": 450000
        }
      ]
    }
  ]
}
```

#### Asset Type Detection

Automatically categorizes assets:
- **PNG**: Image assets (detected by name or extension)
- **PDF**: Vector assets
- **JPEG**: Photo assets
- **Color**: Named colors
- **Data**: Data assets
- **Unknown**: Unclassified assets

#### Scale Detection

Identifies asset scales:
- `@1x`, `@2x`, `@3x` - iOS scale factors
- `tablet`, `compact` - Size classes
- Detected from asset names and metadata

#### Asset Optimization

Generates suggestions for oversized assets:

```json
{
  "category": "assets",
  "severity": "low",
  "title": "Large asset: background-image",
  "description": "Asset is 2.5 MB",
  "files": ["background-image"],
  "impact": 2621440,
  "action": "Consider compressing or resizing asset"
}
```

#### BOM File Structure

Assets.car files use Apple's BOM (Bill of Materials) format:
- Parses BOM headers and blocks
- Reads FACETKEYS tree for asset names
- Handles LZFSE-compressed content (if present)
- Graceful fallback to file size on parse errors

### 4. LZFSE Compression Support

Modern iOS IPAs may use LZFSE compression instead of standard DEFLATE.

#### Automatic Detection

The extractor automatically detects LZFSE compression:
- Checks ZIP compression method (99 = LZFSE)
- Inspects magic bytes: `bvx-`, `bvx$`, `bvxn`, `bvx2`
- Transparently decompresses during extraction

#### LZFSE Magic Bytes

| Magic | Description |
|-------|-------------|
| `bvx-` | LZFSE compressed block |
| `bvx$` | LZFSE end of stream |
| `bvxn` | LZFSE uncompressed block |
| `bvx2` | LZVN compressed block |

#### Decompression

Uses `github.com/blacktop/lzfse-cgo` (Apple's official LZFSE library):
- CGO wrapper for high-performance decompression
- Automatic fallback for standard ZIP formats
- Zero configuration required

#### Error Handling

If LZFSE decompression fails:

```
LZFSE decompression failed for Payload/App.app/Binary: output is empty
Hint: This IPA uses LZFSE compression (method 99). Make sure LZFSE support is enabled.
```

## Usage

### Basic Analysis

```bash
# Analyze iOS app bundle
./bundle-inspector analyze Wikipedia.app -o json

# Analyze IPA (automatically extracts and analyzes)
./bundle-inspector analyze app.ipa -o json
```

### Accessing Advanced Metadata

```bash
# View binary information
./bundle-inspector analyze app.ipa -o json | jq '.metadata.binaries'

# View framework dependencies
./bundle-inspector analyze app.ipa -o json | jq '.metadata.frameworks'

# View asset catalog info
./bundle-inspector analyze app.ipa -o json | jq '.metadata.asset_catalogs'

# View dependency graph
./bundle-inspector analyze app.ipa -o json | jq '.metadata.dependency_graph'
```

### Finding Optimization Opportunities

```bash
# View all optimizations
./bundle-inspector analyze app.ipa -o json | jq '.optimizations'

# Filter by category
./bundle-inspector analyze app.ipa -o json | jq '.optimizations[] | select(.category == "frameworks")'

# Calculate total savings
./bundle-inspector analyze app.ipa -o json | jq '[.optimizations[].impact] | add'
```

## Implementation Details

### Package Structure

```
internal/analyzer/ios/
├── ipa.go                      # IPA analyzer (extracts and analyzes)
├── app.go                      # .app bundle analyzer
├── macho/
│   ├── parser.go              # Mach-O binary parsing
│   ├── architecture.go        # Architecture detection
│   ├── dependencies.go        # Dependency graph construction
│   └── parser_test.go         # Unit tests
├── assets/
│   ├── car_parser.go          # Assets.car parsing
│   ├── asset_info.go          # Asset metadata types
│   └── car_parser_test.go    # Unit tests
└── compression/
    ├── lzfse.go               # LZFSE decompression
    ├── lzfse_test.go          # Unit tests
    └── README.md              # LZFSE documentation
```

### Dependencies

- **Go stdlib**: `debug/macho` for Mach-O parsing
- **howett.net/plist**: Info.plist parsing (framework metadata)
- **github.com/iineva/bom**: BOM/Assets.car file parsing
- **github.com/blacktop/lzfse-cgo**: LZFSE decompression (CGO)

### Performance

All features are designed for minimal performance impact:
- **Mach-O parsing**: Uses efficient stdlib parser, ~10ms per binary
- **Framework discovery**: Single directory walk, parallel binary parsing
- **Assets.car parsing**: Lazy loading, only parses when .car files present
- **LZFSE decompression**: Native C library performance via CGO

### Error Handling

All features implement graceful degradation:
- **Parse failures**: Log warnings, continue analysis
- **Missing files**: Skip gracefully, no crash
- **Unsupported formats**: Fall back to basic file enumeration
- **Warnings stored**: Available in `metadata.analysis_warnings`

## Examples

### Real-World Analysis: Wikipedia.app

Analyzing the Wikipedia iOS app demonstrates all features:

**Binary Analysis:**
- Main binary: `Wikipedia` (arm64 executable, 16 KB)
- Framework: `WMF.framework/WMF` (arm64 dylib, 10.1 MB code)
- Debug symbols: `Wikipedia.debug.dylib` (23.8 MB with DWARF)
- Extensions: Multiple app extensions with binaries

**Framework Dependencies:**
```
Wikipedia → Wikipedia.debug.dylib → WMF.framework
ContinueReadingWidget → ContinueReadingWidget.debug.dylib → WMF.framework
```

**Assets:**
- Main catalog: 331 assets, 2.9 MB
- Extension catalogs: 4 additional Assets.car files
- Top asset: AppIcon (various scales)

**Optimizations Detected:**
- Large debug dylib (23.8 MB) - strip for production
- Duplicate GIF files across extensions
- Multiple Assets.car files (deduplicate opportunity)

## Testing

### Test Artifacts

The project uses real iOS applications for testing:

- **Wikipedia.app**: Full-featured iOS app with frameworks, extensions, assets
- **lightyear.ipa**: Large production IPA (81 MB compressed)

### Running Tests

```bash
# Unit tests (all iOS analyzers)
go test ./internal/analyzer/ios/...

# Specific package tests
go test ./internal/analyzer/ios/macho/... -v
go test ./internal/analyzer/ios/assets/... -v
go test ./internal/analyzer/ios/compression/... -v

# Integration test with real artifact
./bundle-inspector analyze test-artifacts/ios/Wikipedia.app -o json
```

### Test Coverage

- **Mach-O parsing**: 9 tests (architecture, libraries, segments, debug symbols)
- **Framework analysis**: 8 tests (discovery, dependency graph, unused detection)
- **Assets.car parsing**: 8 tests (extraction, categorization, type/scale detection)
- **LZFSE compression**: 3 tests (magic bytes, decompression, error handling)
- **Total**: 33+ unit tests, all passing ✅

## Future Enhancements

Potential improvements for future phases:

1. **Enhanced Asset Analysis**
   - Image size analysis (actual pixel dimensions)
   - Compression ratio detection
   - Unused asset detection

2. **Advanced Framework Analysis**
   - Weak vs strong linking detection
   - Framework code signing analysis
   - Framework versioning conflicts

3. **Bitcode Analysis**
   - Bitcode section detection
   - Size impact calculation

4. **Symbols Analysis**
   - Symbol table parsing
   - Exported symbols enumeration
   - Undefined symbols detection

5. **Entitlements**
   - Parse embedded.mobileprovision
   - Extract entitlements plist
   - Capability analysis

## References

- [Mach-O File Format](https://github.com/aidansteele/osx-abi-macho-file-format-reference)
- [Apple's LZFSE](https://github.com/lzfse/lzfse)
- [Assets.car Format](https://blog.timac.org/2018/1018-reverse-engineering-the-car-file-format/)
- [BOM File Format](https://github.com/hogliux/bomutils)
- [iOS App Structure](https://www.appdome.com/how-to/devsecops-automation-mobile-cicd/appdome-basics/structure-of-an-ios-app-binary-ipa/)
