# Test Artifacts

This directory contains real mobile applications used for integration testing and validation of the bundle-inspector tool.

## Artifacts

### iOS

#### 1. lightyear.ipa (81 MB)
- **Type:** iOS App Archive
- **Format:** IPA (ZIP-based)
- **Purpose:** Test IPA extraction, iOS-specific analysis
- **Expected Features:**
  - Compressed size: ~81MB
  - Contains .app bundle
  - Frameworks, resources, assets
  - Executable binary

#### 2. Wikipedia.app
- **Type:** iOS App Bundle (uncompressed)
- **Format:** Directory structure
- **Size:** ~24MB (debug dylib)
- **Purpose:** Test direct .app analysis without IPA wrapper
- **Notable Contents:**
  - Large debug symbols (Wikipedia.debug.dylib - 24MB)
  - Assets.car (2.8MB)
  - 100+ localizations (.lproj directories)
  - Multiple frameworks
  - NIB/storyboard files

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

## Notes

- **Large Files:** These artifacts are excluded from git (see `.gitignore`)
- **Local Testing Only:** Artifacts are for local development and testing
- **Privacy:** Ensure artifacts don't contain sensitive data before committing
- **JSON Reports:** Analysis generates JSON reports in the same directory
