# Bitrise Bundle Inspector Plugin

A Bitrise CLI plugin that analyzes mobile artifacts (iOS and Android) for size optimization opportunities. Inspired by [Tuist Inspect](https://tuist.dev/blog/2025/05/15/bundle-size-analysis), [Emerge Tools](https://www.emergetools.com/product/sizeanalysis), and [Expo Atlas](https://docs.expo.dev/guides/analyzing-bundles/).

## Features

### Current (MVP - Phases 1-3 Complete) ✅
- ✅ iOS IPA analysis with size breakdown
- ✅ Android APK/AAB analysis
- ✅ Recursive bundle traversal
- ✅ File enumeration and categorization
- ✅ Top largest files identification
- ✅ Duplicate file detection with SHA-256 hashing
- ✅ Large file detection with configurable threshold
- ✅ Optimization recommendations with severity levels
- ✅ Human-readable text output
- ✅ JSON output for CI/CD integration

### Coming Soon
- 🚧 iOS advanced features (Mach-O parsing, LZFSE, Assets.car) (Phase 4)
- 🚧 HTML interactive reports (Phase 5)
- 🚧 Artifact comparison (Phase 5)
- 🚧 Full Android manifest parsing (Phase 4)

## Supported Artifact Types

- **iOS**: `.ipa`, `.app`, `.xcarchive` (IPA currently implemented)
- **Android**: `.apk`, `.aab` (coming in Phase 2)

## Installation

### From Source (Development)

```bash
# Clone the repository
git clone https://github.com/bitrise-io/bitrise-plugins-bundle-inspector.git
cd bitrise-plugins-bundle-inspector

# Build the CLI
go build -o bundle-inspector ./cmd/bundle-inspector

# Optional: Install to PATH
go install ./cmd/bundle-inspector
```

### As Bitrise Plugin (Future)

Once released, install via Bitrise CLI:

```bash
bitrise plugin install https://github.com/bitrise-io/bitrise-plugins-bundle-inspector.git
```

## Usage

### Analyze an Artifact

```bash
# Analyze an iOS IPA
./bundle-inspector analyze path/to/app.ipa

# Analyze and save to file
./bundle-inspector analyze path/to/app.ipa -f report.txt
```

### Output Formats

- `text` (default): Human-readable format
- `json`: Machine-parseable format for CI/CD integration
- `html`: Interactive web report (coming in Phase 5)

```bash
# JSON output
./bundle-inspector analyze app.ipa -o json -f report.json

# Pretty-printed JSON to stdout
./bundle-inspector analyze app.apk -o json

# HTML report (coming soon)
./bundle-inspector analyze app.ipa -o html -f report.html
```

### Command Reference

```bash
# Analyze a single artifact
bundle-inspector analyze <file-path> [flags]

# Compare two artifacts (coming in Phase 5)
bundle-inspector compare <file1> <file2> [flags]

# Show version information
bundle-inspector version
```

#### Flags

- `--output, -o`: Output format (text/json/html) [default: text]
- `--output-file, -f`: Write output to file instead of stdout
- `--threshold`: Large file warning threshold [default: "1MB"]
- `--include-duplicates`: Enable duplicate detection [default: true]
- `--fail-on-increase`: Exit with error if size increases (compare command, Phase 5)

## Example Output

```
Bundle Inspector Analysis Report
=================================

Artifact Information:
  Type: ipa
  Path: MyApp.ipa
  Compressed Size: 45.2 MB
  Uncompressed Size: 52.8 MB
  Compression Ratio: 85.6%

Size Breakdown:
  Executable: 15.3 MB (29.0%)
  Frameworks: 22.1 MB (41.9%)
  Assets: 8.4 MB (15.9%)
  Resources: 5.2 MB (9.8%)
  Other: 1.8 MB (3.4%)

Top 10 Largest Files:
   1. Frameworks/MyFramework.framework/MyFramework - 12.5 MB (23.7%)
   2. MyApp - 15.3 MB (29.0%)
   3. Assets.car - 6.2 MB (11.7%)
   ...

Duplicate Files:
  2 copies of image.png (1.2 MB each):
    - Assets/image.png
    - Resources/image.png
    Wasted space: 1.2 MB
  Total wasted space: 1.2 MB

Optimization Opportunities:

  High Priority:
    • Remove 1 duplicate copies of files
      Found 2 identical files (1.2 MB each)
      Potential savings: 1.2 MB
      Action: Keep only one copy and deduplicate references

Total Potential Savings: 1.2 MB (2.3%)
```

## Development

### Project Structure

```
bundle-inspector/
├── cmd/bundle-inspector/     # CLI entry point
├── internal/
│   ├── analyzer/             # Artifact analyzers
│   │   ├── ios/              # iOS analyzers
│   │   └── android/          # Android analyzers
│   ├── detector/             # Duplicate/bloat detection
│   ├── report/               # Output formatters
│   └── util/                 # Utilities
├── pkg/types/                # Public API types
└── bitrise-plugin.yml        # Plugin configuration
```

### Building

```bash
# Build for current platform
go build -o bundle-inspector ./cmd/bundle-inspector

# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Testing with Sample Artifacts

```bash
# Analyze a test IPA
./bundle-inspector analyze test-artifacts/sample.ipa

# Test error handling
./bundle-inspector analyze non-existent.ipa
./bundle-inspector analyze invalid.zip
```

## Implementation Roadmap

### Phase 1: Foundation ✅ COMPLETE
- [x] Go module initialization
- [x] Project structure
- [x] Cobra CLI setup
- [x] iOS IPA analyzer
- [x] Text output formatter
- [x] Basic README

### Phase 2: Android Support ✅ COMPLETE
- [x] APK analyzer
- [x] AAB analyzer
- [x] DEX file enumeration and categorization
- [x] Native library detection by architecture
- [x] Module detection for AAB

### Phase 3: Advanced Detection ✅ COMPLETE
- [x] SHA-256 duplicate detection with parallel hashing
- [x] Large file/bloat detection with configurable threshold
- [x] Optimization recommendations with severity levels
- [x] JSON output formatter

### Phase 4: iOS Advanced Features
- [ ] Mach-O binary parsing
- [ ] LZFSE compression support
- [ ] Assets.car parsing
- [ ] Framework dependency analysis

### Phase 5: Comparison & HTML
- [ ] compare command
- [ ] Delta calculations
- [ ] HTML report generator
- [ ] Interactive visualizations

### Phase 6: Release
- [ ] Comprehensive testing
- [ ] Performance benchmarks
- [ ] GoReleaser automation
- [ ] GitHub Actions CI/CD
- [ ] Official plugin registration

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

Copyright © Bitrise

## References

- [Tuist Bundle Size Analysis](https://tuist.dev/blog/2025/05/15/bundle-size-analysis)
- [Emerge Tools Size Analysis](https://www.emergetools.com/product/sizeanalysis)
- [Expo Atlas](https://docs.expo.dev/guides/analyzing-bundles/)
- [iOS App Binary Structure](https://www.appdome.com/how-to/devsecops-automation-mobile-cicd/appdome-basics/structure-of-an-ios-app-binary-ipa/)
- [Android App Bundle Format](https://developer.android.com/guide/app-bundle/app-bundle-format)
