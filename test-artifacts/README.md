# Test Artifacts

This directory contains real mobile applications used for integration testing and validation of the bundle-inspector tool.

## Artifacts

### iOS

#### 1. lightyear.ipa (81 MB)
- **Type:** iOS App Archive
- **Format:** IPA (ZIP-based)
- **Compression:** Standard DEFLATE (not LZFSE)
- **Purpose:** Test IPA extraction, iOS-specific analysis, large file handling
- **Expected Features:**
  - Compressed size: ~81MB, uncompressed: ~124MB
  - Contains Runner.app bundle
  - 115MB of frameworks (multiple architectures)
  - Assets.car files
  - Executable binary (7MB)
  - 3175 files total

**Note:** This IPA uses standard ZIP compression (DEFLATE). LZFSE compression support (method 99) is tested when modern IPAs with LZFSE are analyzed.

#### 2. Wikipedia.app
- **Type:** iOS App Bundle (uncompressed)
- **Format:** Directory structure
- **Size:** ~145MB uncompressed
- **Purpose:** Test direct .app analysis without IPA wrapper, Phase 4 advanced features testing
- **Notable Contents:**
  - Large debug symbols (Wikipedia.debug.dylib - 23.8MB)
  - Assets.car (2.9MB with 331 assets)
  - 100+ localizations (.lproj directories)
  - WMF.framework (34.2MB)
  - Multiple app extensions (PlugIns/)
  - NIB/storyboard files

**Phase 4 Advanced Features Detected:**
- **Mach-O Binaries:**
  - Main executable: `Wikipedia` (arm64, 16KB executable)
  - Debug dylib: `Wikipedia.debug.dylib` (23.8MB, DWARF symbols)
  - Framework binary: `Frameworks/WMF.framework/WMF` (arm64 dylib, 10.1MB code)
  - Extension binaries: Multiple app extensions with arm64 binaries
- **Framework Analysis:**
  - WMF.framework v7.8.1 (34.2MB total)
  - Automatic version detection from Info.plist
  - Framework dependencies to system frameworks (Foundation, UIKit, etc.)
  - Dependency graph showing Wikipedia → WMF linkage
- **Assets.car Parsing:**
  - Main Assets.car: 331 assets, 2.9MB
  - Extension catalogs: 4 additional Assets.car files in PlugIns/
  - Asset types: PNG, PDF, data assets
  - Notable assets: AppIcon, W logo, activity icons
- **Dependency Graph:**
  ```
  Wikipedia → Wikipedia.debug.dylib → WMF.framework
  ContinueReadingWidget → ContinueReadingWidget.debug.dylib → WMF.framework
  NotificationServiceExtension → NotificationServiceExtension.debug.dylib → WMF.framework
  WidgetsExtension → WidgetsExtension.debug.dylib → WMF.framework
  ```

### Android

#### 3. 2048-game-2048.apk (11 MB)
- **Type:** Android Application Package
- **Format:** APK (ZIP-based)
- **Purpose:** Test APK analysis, DEX detection, Android resources
- **Expected Features:**
  - DEX files
  - Native libraries (if multi-arch)
  - Resources (res/ directory)
  - Assets
  - AndroidManifest.xml

## Usage

### Run All Analyses
```bash
./scripts/analyze-all-test-artifacts.sh
```

### Run Integration Tests
```bash
./scripts/run-integration-tests.sh
```

### Analyze Individual Artifacts
```bash
# iOS IPA
./bundle-inspector analyze test-artifacts/ios/lightyear.ipa

# iOS App Bundle
./bundle-inspector analyze test-artifacts/ios/Wikipedia.app

# Android APK
./bundle-inspector analyze test-artifacts/android/2048-game-2048.apk
```

## Phase 4 Advanced Features Testing

### Mach-O Binary Parsing
Test with Wikipedia.app to verify:
```bash
./bundle-inspector analyze test-artifacts/ios/Wikipedia.app -o json | jq '.metadata.binaries'
```

Expected output includes:
- Architecture detection (arm64)
- Binary types (executable, dylib)
- Code/data segment sizes
- Linked libraries with @rpath references
- Debug symbols detection

### Framework Dependency Analysis
```bash
./bundle-inspector analyze test-artifacts/ios/Wikipedia.app -o json | jq '.metadata.frameworks'
```

Expected output includes:
- WMF.framework discovery
- Version information (v7.8.1)
- Framework size calculation
- Binary architecture details
- System framework dependencies

### Dependency Graph
```bash
./bundle-inspector analyze test-artifacts/ios/Wikipedia.app -o json | jq '.metadata.dependency_graph'
```

Expected output shows:
- Main app → debug dylib → WMF framework
- Extension → extension debug dylib → WMF framework
- Complete transitive dependency closure

### Assets.car Parsing
```bash
./bundle-inspector analyze test-artifacts/ios/Wikipedia.app -o json | jq '.metadata.asset_catalogs'
```

Expected output includes:
- Multiple Assets.car files (main app + extensions)
- Asset counts (331 assets in main catalog)
- Type categorization (PNG, PDF, data)
- Largest assets identification
- Total catalog sizes

### LZFSE Compression
LZFSE support is automatically enabled. When analyzing IPAs with LZFSE compression (method 99):
```bash
./bundle-inspector analyze modern-app.ipa
```

The tool will:
- Automatically detect compression method 99
- Decompress with LZFSE library
- Extract and analyze transparently
- Show helpful error if decompression fails

## Notes

- **Large Files:** These artifacts are excluded from git (see `.gitignore`)
- **Local Testing Only:** Artifacts are for local development and testing
- **Privacy:** Ensure artifacts don't contain sensitive data before committing
- **JSON Reports:** Analysis generates JSON reports in the same directory
- **Phase 4 Testing:** Wikipedia.app is the primary test artifact for iOS advanced features
