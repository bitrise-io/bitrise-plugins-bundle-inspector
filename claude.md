# Claude Context: Bitrise Bundle Inspector Plugin

## Project Overview

Bundle Inspector is a **Bitrise CLI plugin** that analyzes mobile app artifacts (iOS and Android) for size optimization opportunities. It's designed primarily for CI/CD integration to help teams monitor, track, and optimize app bundle sizes.

**Primary Use Case**: Bitrise CI/CD workflows (80% of users)
**Secondary Use Cases**: Local testing (15%), Development (5%)

## Key Value Propositions

1. **Auto-detection**: Automatically finds artifacts from Bitrise environment variables
2. **Multiple Output Formats**: Text, JSON, Markdown, HTML (with interactive treemaps)
3. **Automatic Export**: Reports sent to Bitrise deploy directory for easy access
4. **Deep Analysis**: iOS Mach-O parsing, framework dependencies, Assets.car analysis, Android DEX class-level parsing
5. **Optimization Recommendations**: Actionable suggestions with severity levels

## Architecture

### Language & Framework
- **Language**: Go 1.21+
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **Build System**: Go modules

### Directory Structure

```
bundle-inspector/
├── cmd/bundle-inspector/     # CLI entry point (main.go)
├── internal/
│   ├── analyzer/             # Platform-specific analyzers
│   │   ├── ios/              # IPA, .app, XCArchive analyzers
│   │   │   ├── macho/        # Mach-O binary parser
│   │   │   ├── assetcar/     # Assets.car parser
│   │   │   └── framework/    # Framework dependency analyzer
│   │   └── android/          # APK, AAB analyzers
│   │       └── dex/          # DEX file parser (class-level analysis)
│   ├── bitrise/              # Bitrise CI integration (env.go)
│   ├── detector/             # Duplicate & bloat detection
│   ├── orchestrator/         # Analysis coordination
│   ├── report/               # Output formatters (text, json, markdown, html)
│   └── util/                 # Utilities (humanize, zip)
├── pkg/types/                # Public API types
├── test-artifacts/           # Real mobile apps for testing (NOT in git)
├── scripts/                  # Helper scripts
└── bitrise-plugin.yml        # Plugin configuration
```

### Key Files

#### Entry Point
- **cmd/bundle-inspector/main.go**: CLI entry, Cobra commands, flag parsing, output handling

#### Bitrise Integration
- **internal/bitrise/env.go**: Core integration logic
  - `DetectBundlePath()`: Auto-detects artifacts (IPA → AAB → APK priority)
  - `GetBuildMetadata()`: Reads Bitrise env vars
  - `ExportToDeployDir()`: Copies reports to `$BITRISE_DEPLOY_DIR`

#### Analyzers
- **internal/analyzer/ios/ipa.go**: IPA file analyzer
- **internal/analyzer/ios/app.go**: .app bundle analyzer
- **internal/analyzer/ios/xcarchive.go**: XCArchive analyzer
- **internal/analyzer/ios/macho/**: Mach-O binary parsing (architecture, libraries, symbols)
- **internal/analyzer/ios/assetcar/**: Assets.car parsing (asset extraction, categorization)
- **internal/analyzer/ios/framework/**: Framework dependency analysis
- **internal/analyzer/android/apk.go**: APK analyzer
- **internal/analyzer/android/aab.go**: AAB analyzer

#### Detection
- **internal/detector/duplicate.go**: SHA-256 duplicate file detection (parallel)
- **internal/detector/duplicate_categorizer.go**: Intelligent filtering orchestrator
- **internal/detector/path_analysis.go**: Bundle/framework/SDK path analysis utilities
- **internal/detector/rules.go**: Rule registry and filtering framework
- **internal/detector/rule_info_plist.go**: Rule 1 - Info.plist bundle boundary detection
- **internal/detector/rule_nib_variants.go**: Rule 2 - NIB version variants detection
- **internal/detector/rule_contents_json.go**: Rule 3 - Asset catalog Contents.json detection
- **internal/detector/rule_localization.go**: Rule 4 - Localization file bundle isolation
- **internal/detector/rule_framework_scripts.go**: Rule 5 - Framework build scripts detection
- **internal/detector/rule_framework_metadata.go**: Rule 6 - Framework metadata files detection
- **internal/detector/rule_third_party_sdk.go**: Rule 7 - Third-party SDK resources (100+ SDKs)
- **internal/detector/rule_extension_duplication.go**: Rule 8 - Extension resource duplication (actionable)
- **internal/detector/rule_asset_duplication.go**: Rule 9 - Asset duplication in same bundle (actionable)
- **internal/detector/bloat.go**: Large file detection

#### Output Formatters
- **internal/report/text.go**: Human-readable console output
- **internal/report/json.go**: Machine-parseable JSON
- **internal/report/markdown.go**: GitHub/GitLab-friendly markdown
- **internal/report/html.go**: Interactive HTML with D3.js treemaps

#### Types
- **pkg/types/types.go**: Public API types (Report, ArtifactInfo, SizeBreakdown, Duplicate, Optimization)

## Core Workflows

### 1. Analysis Flow
```
User runs command
  ↓
CLI detects artifact path (explicit or auto-detect)
  ↓
Orchestrator determines artifact type
  ↓
Appropriate analyzer runs (iOS or Android)
  ↓
Detectors run (duplicates, bloat)
  ↓
Optimization recommendations generated
  ↓
Report formatted (text/json/markdown/html)
  ↓
Report written to file
  ↓
If Bitrise: Export to $BITRISE_DEPLOY_DIR
```

### 2. Auto-Detection Flow
```
No explicit path provided
  ↓
Check $BITRISE_IPA_PATH → if exists, use it
  ↓
Check $BITRISE_AAB_PATH → if exists, use it
  ↓
Check $BITRISE_APK_PATH → if exists, use it
  ↓
If none found → error
```

### 3. Bitrise Export Flow
```
Analysis complete
  ↓
Check if $BITRISE_DEPLOY_DIR is set
  ↓
If yes:
  - Export JSON report → bundle-analysis.json
  - Export text report → bundle-report.txt
  - Export markdown report → bundle-analysis.md (best-effort)
  ↓
Reports appear in Build Artifacts tab
```

## Environment Variables

### Bitrise Environment Variables (Used by Plugin)

| Variable | Purpose | Used For |
|----------|---------|----------|
| `BITRISE_IPA_PATH` | iOS IPA path | Auto-detection (priority 1) |
| `BITRISE_AAB_PATH` | Android AAB path | Auto-detection (priority 2) |
| `BITRISE_APK_PATH` | Android APK path | Auto-detection (priority 3) |
| `BITRISE_DEPLOY_DIR` | Deploy directory | Automatic report export |
| `BITRISE_BUILD_NUMBER` | Build number | Report metadata |
| `GIT_CLONE_COMMIT_HASH` | Git commit | Report metadata |

### Detection Priority
IPA (1st) → AAB (2nd) → APK (3rd)

## Command-Line Interface

### Commands
```bash
bundle-inspector analyze [file-path] [flags]  # Analyze artifact
bundle-inspector version                       # Show version info
```

### Flags
- `-o, --output`: Output format(s) - comma-separated for multiple (text/json/markdown/html) [default: text]
- `-f, --output-file`: Output filename(s) - comma-separated when using multiple formats (default: auto-generated)
- `--include-duplicates`: Enable duplicate detection [default: true]
- `--no-auto-detect`: Disable Bitrise auto-detection

### Usage Patterns
```bash
# As Bitrise plugin (recommended)
bitrise :bundle-inspector analyze

# Single format
./bundle-inspector analyze path/to/app.ipa -o json -f report.json

# Multiple formats (efficient - runs analysis once)
./bundle-inspector analyze path/to/app.ipa -o json,markdown,html

# Multiple formats with custom filenames
./bundle-inspector analyze path/to/app.ipa -o json,html -f data.json,report.html

# All 4 formats at once
./bundle-inspector analyze path/to/app.ipa -o text,json,markdown,html
```

## Output Formats

### 1. Text (Default)
- Human-readable console output
- Sections: Artifact Info, Size Breakdown, Top Files, Duplicates, Optimizations
- Best for: Quick inspection, CI logs

### 2. JSON
- Complete structured data
- Machine-parseable
- Best for: Automation, size checks, API integration
- Example: `jq '.artifact_info.size' report.json`

### 3. Markdown
- GitHub-flavored markdown
- Tables and formatted sections
- Best for: PR comments, documentation, Slack
- Example: `gh pr comment $PR_NUMBER --body-file report.md`

### 4. HTML
- Interactive D3.js treemap visualization
- Sortable tables, collapsible sections
- Best for: Stakeholder reports, exploration

## Testing

### Test Artifacts (NOT in Git)
```
test-artifacts/
├── ios/
│   ├── lightyear.ipa          # 81MB iOS game
│   └── Wikipedia.app/         # Real .app bundle
└── android/
    └── 2048-game-2048.apk     # 11MB Android game
```

### Running Tests
```bash
go test ./...                   # All unit tests
go test -cover ./...            # With coverage
./scripts/run-integration-tests.sh  # Integration tests

# Manual testing
./bundle-inspector analyze test-artifacts/ios/lightyear.ipa
```

## Common Development Tasks

### Building
```bash
go build -o bundle-inspector ./cmd/bundle-inspector
```

### Adding a New Analyzer
1. Create analyzer in `internal/analyzer/{platform}/`
2. Implement analyzer interface
3. Add type detection in `orchestrator/orchestrator.go`
4. Add tests
5. Update documentation

### Adding a New Output Format
1. Create formatter in `internal/report/{format}.go`
2. Implement `Format(io.Writer, *types.Report) error`
3. Add case in `cmd/bundle-inspector/main.go` → `writeReport()`
4. Add tests
5. Update README.md

### Adding a New Detection Algorithm
1. Create detector in `internal/detector/{name}.go`
2. Add to orchestrator workflow
3. Add optimization recommendation generation
4. Add tests

## Code Conventions

### Error Handling
- Use `fmt.Errorf("message: %w", err)` for wrapping
- Return errors, don't panic
- Log warnings to stderr with `fmt.Fprintf(os.Stderr, ...)`

### Logging
- Info/Progress: Write to `os.Stderr` (doesn't pollute stdout)
- Reports: Write to `os.Stdout` or files
- Format: `fmt.Fprintf(os.Stderr, "✓ Action completed: %s\n", result)`

### File Paths
- Always use absolute paths
- Use `filepath` package for cross-platform compatibility
- Validate paths with `os.Stat()` before use

### Parallelization
- Use goroutines for independent operations (e.g., file hashing)
- Use `sync.WaitGroup` for coordination
- Example: Duplicate detection uses parallel hashing

## Important Implementation Details

### iOS Analysis
- **LZFSE Compression**: Automatically handles compression method 99 (modern iOS IPAs)
- **Mach-O Parsing**: Detects architecture (arm64, x86_64), linked libraries, code/data sizes
- **Framework Discovery**: Reads Info.plist, analyzes dependencies, detects unused frameworks
- **Assets.car**: Extracts assets, categorizes by type (@1x, @2x, @3x) and format

### Android Analysis
- **APK**: ZIP-based analysis, DEX class-level parsing, native library detection
- **AAB**: Module detection, base/dynamic features, DEX class-level parsing
- **DEX Parsing**:
  - Parses all DEX files (classes.dex, classes2.dex, ...) and merges into unified virtual "Dex/" directory
  - Class-level breakdown with package hierarchy (e.g., Dex/com/example/app/MainActivity.class)
  - Private size calculation: size 100% attributable to each class (methods, fields, annotations)
  - _Unmapped node: shows shared data structures (string pools, type descriptors) that can't be attributed to specific classes
  - Obfuscation detection: identifies when >50% of classes have single-letter names
  - Metadata tracking: method count, field count, source DEX file per class
  - Uses dextk library (github.com/csnewman/dextk) for pure-Go DEX parsing
- **Architecture Detection**: Groups .so files by architecture (arm64-v8a, armeabi-v7a, x86, x86_64)

### Duplicate Detection with Intelligent Filtering ⭐

**Detection Phase:**
- SHA-256 hashing for file identity
- Parallel processing with worker pools
- Groups duplicates by hash
- Calculates wasted space (size × (count - 1))

**Filtering Phase (NEW in v0.3.0):**
- 9-rule filtering system eliminates false positives
- Rules 1-7: Filter architectural patterns and third-party SDK resources
- Rules 8-9: Identify actionable duplicates with priority (high/medium/low)
- **Result**: 60-80% reduction in false positives

**Rule-Based Categorization:**
1. **Rule 1**: Info.plist in different bundles → FILTER (iOS requirement)
2. **Rule 2**: NIB version variants → FILTER (iOS compatibility)
3. **Rule 3**: Contents.json in asset catalogs → FILTER (required metadata)
4. **Rule 4**: Localization files in different bundles → FILTER (bundle isolation)
5. **Rule 5**: Framework build scripts → FILTER (CocoaPods/Carthage artifacts)
6. **Rule 6**: Framework metadata (.supx, .bcsymbolmap, etc.) → FILTER (required metadata)
7. **Rule 7**: Third-party SDK resources (100+ SDKs) → FILTER (not under developer control)
8. **Rule 8**: Extension resource duplication → ACTIONABLE with priority
9. **Rule 9**: Asset duplication in same bundle → ACTIONABLE with priority

**Architecture:**
```
Duplicate Detection
    ↓
DuplicateCategorizer.EvaluateDuplicate()
    ↓
Rule Evaluation (9 rules)
    ↓
FilterResult (ShouldFilter=true/false, Priority)
    ↓
Orchestrator: Skip if filtered, create optimization if actionable
    ↓
Only actionable optimizations in report
```

**Key Files:**
- `internal/detector/duplicate.go`: SHA-256 detection
- `internal/detector/duplicate_categorizer.go`: Categorization orchestrator
- `internal/detector/path_analysis.go`: Bundle/framework/SDK detection
- `internal/detector/rules.go`: Rule registry
- `internal/detector/rule_*.go`: Individual filtering rules (9 files)
- `internal/orchestrator/orchestrator.go`: Integration point (generateOptimizations)

**See**: `docs/duplicate-detection.md` for complete rule documentation

### Optimization Recommendations
- Severity levels: High, Medium, Low
- Each recommendation includes:
  - Title (what to do)
  - Description (why)
  - Potential savings (bytes)
  - Action (how to fix)

### Multi-Format Output
- **Single Analysis, Multiple Formats**: Analysis runs once, generates multiple report formats
- **Comma-Separated Formats**: `-o json,markdown,html` generates all three formats
- **Custom Filenames**: `-f file1.json,file2.md,file3.html` (must match format count)
- **Auto-Generated Names**: Default: `bundle-analysis-<artifact>.<ext>`
- **Benefits**:
  - Faster (no re-analysis for large artifacts)
  - Consistent (all formats from same analysis run)
  - Convenient (one command for multiple audiences)

## Documentation

### User Documentation
- **README.md**: User-facing docs (Bitrise users, CI/CD focus)
- **QUICKSTART.md**: Developer Guide (building, testing, contributing)
- **docs/ios-advanced-analysis.md**: Deep dive on iOS features
- **docs/duplicate-detection.md**: Intelligent filtering system explained (all 9 rules, examples, FAQ)

### Code Documentation
- Exported functions have doc comments
- Complex logic has inline comments
- Package-level docs at top of key files

## Bitrise Plugin Configuration

### bitrise-plugin.yml
```yaml
name: bundle-inspector
description: Analyze mobile bundles for size optimization
executable:
  osx: https://github.com/.../bundle-inspector-Darwin-x86_64
  osx-arm64: https://github.com/.../bundle-inspector-Darwin-arm64
  linux: https://github.com/.../bundle-inspector-Linux-x86_64
requirements:
  - tool: bitrise
    min_version: 1.3.0
```

### Plugin Installation
```bash
bitrise plugin install https://github.com/bitrise-io/bitrise-plugins-bundle-inspector.git
```

## Common Patterns

### Reading Bitrise Environment
```go
import "github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/bitrise"

if bitrise.IsBitriseEnvironment() {
    metadata := bitrise.GetBuildMetadata()
    // Use metadata.BuildNumber, metadata.CommitHash, metadata.DeployDir
}
```

### Auto-Detecting Artifacts
```go
path, err := bitrise.DetectBundlePath()
if err != nil {
    // No artifact found in environment
}
```

### Exporting to Deploy Directory
```go
destPath, err := bitrise.WriteToDeployDir("report.json", jsonData)
// File copied to $BITRISE_DEPLOY_DIR/report.json
```

## Edge Cases & Gotchas

### 1. LZFSE Compression
- Modern iOS IPAs use compression method 99 (LZFSE)
- Standard Go `archive/zip` doesn't support this
- We use custom decompression via `blacktop/lzfse-cgo`

### 2. Auto-Detection Priority
- IPA takes priority over AAB/APK
- If multiple artifacts exist, first one found wins
- Users can override with explicit path

### 3. Deploy Directory Export
- Always logs to stderr (not stdout)
- Best-effort for markdown export (doesn't fail if errors)
- Creates directory if doesn't exist

### 4. Duplicate Detection
- Can be slow for large artifacts (>100MB)
- Disabled with `--include-duplicates=false`
- Uses SHA-256 (cryptographically secure)

### 5. File Paths in Archives
- iOS: Uses `/` separator (Unix-style)
- Android: Also uses `/` separator (ZIP standard)
- Cross-platform compatible with `filepath.ToSlash()`

## Release Process

This project uses **goreleaser** to automate the release process, building binaries for multiple platforms and publishing to GitHub.

### Prerequisites

- `goreleaser` installed (`brew install goreleaser`)
- GitHub CLI authenticated (`gh auth login`)
- Clean git working directory
- All tests passing

### Release Steps

#### 1. Update Version

Edit `cmd/bundle-inspector/main.go` and update the version constant:

```go
var (
    version = "0.2.0"  // Update this
    commit  = "none"
    date    = "unknown"
)
```

#### 2. Update Plugin Configuration

**IMPORTANT:** Update `bitrise-plugin.yml` with the new version number in all executable URLs:

```yaml
executable:
  osx: https://github.com/bitrise-io/bitrise-plugins-bundle-inspector/releases/download/0.2.0/bundle-inspector-Darwin-x86_64
  osx-arm64: https://github.com/bitrise-io/bitrise-plugins-bundle-inspector/releases/download/0.2.0/bundle-inspector-Darwin-arm64
  linux: https://github.com/bitrise-io/bitrise-plugins-bundle-inspector/releases/download/0.2.0/bundle-inspector-Linux-x86_64
```

Replace `0.2.0` with your new version number. This is **required** - without this update, Bitrise plugin installation will fail with a 404 error.

#### 3. Commit Version Bump

```bash
git add cmd/bundle-inspector/main.go bitrise-plugin.yml
git commit -m "chore: bump version to 0.2.0"
```

#### 4. Ensure Clean Git State

goreleaser requires a clean working directory. Check and clean up:

```bash
# Check status
git status

# If there are uncommitted changes:
# - Add important files to .gitignore (bin/, *.local.json, etc.)
# - Commit or stash changes
# - Remove deleted files: git rm <file>

# Example cleanup:
git add .gitignore
git rm obsolete-file.md
git commit -m "chore: cleanup before release"
```

#### 5. Push Changes

```bash
git push origin main
```

#### 6. Create Git Tag

Create an annotated tag with release notes:

```bash
git tag -a 0.2.0 -m "Release 0.2.0

Brief description of what's in this release.

New features:
- Feature 1
- Feature 2

Bug fixes:
- Fix 1

Technical changes:
- Change 1
"
```

#### 7. Run goreleaser

Export GitHub token and run goreleaser:

```bash
export GITHUB_TOKEN=$(gh auth token)
goreleaser release --clean
```

This will:
- Run `go mod tidy`
- Run all tests (`go test ./...`)
- Build binaries for:
  - Darwin (macOS) ARM64
  - Darwin (macOS) x86_64
  - Linux x86_64
- Inject version, commit, and date into binaries via ldflags
- Generate SHA256 checksums
- Create GitHub release
- Upload all binaries as release assets

#### 8. Verify Release

```bash
# View release details
gh release view 0.2.0

# Check assets
gh release view 0.2.0 --json assets --jq '.assets[].name'

# Open in browser
gh release view 0.2.0 --web
```

### goreleaser Configuration

The release configuration is defined in `.goreleaser.yml`:

**Key settings:**
- **Binary format**: Raw binaries (not archives) for Bitrise compatibility
- **Platforms**: Darwin (arm64, x86_64), Linux (x86_64)
- **Naming**: `bundle-inspector-{OS}-{Arch}`
- **Version injection**: Via ldflags into `main.version`, `main.commit`, `main.date`
- **Release title**: `{Version}` (version number only, no "v" prefix)
- **Changelog**: Auto-generated, excludes `docs:`, `test:`, `chore:`, `ci:` commits

### Troubleshooting

**Issue: "git is in a dirty state"**
```bash
# Check what's dirty
git status

# Clean up:
git add .gitignore  # Add ignored files
git rm deleted-file.md  # Remove deleted files
git restore modified-file  # Restore unwanted changes
git commit -m "chore: cleanup"
```

**Issue: "missing GITHUB_TOKEN"**
```bash
# Ensure GitHub CLI is authenticated
gh auth status

# Export token
export GITHUB_TOKEN=$(gh auth token)
```

**Issue: Tests failing**
```bash
# Run tests locally first
go test ./...

# Fix failing tests before releasing
```

**Issue: Tag already exists**
```bash
# Delete local and remote tag
git tag -d 0.2.0
git push origin :0.2.0

# Recreate with new commit
git tag -a 0.2.0 -m "Release 0.2.0"
```

### Release Artifacts

Each release includes:

| File | Description | Size |
|------|-------------|------|
| `bundle-inspector-Darwin-arm64` | macOS Apple Silicon binary | ~5 MB |
| `bundle-inspector-Darwin-x86_64` | macOS Intel binary | ~5 MB |
| `bundle-inspector-Linux-x86_64` | Linux x86_64 binary | ~5 MB |
| `checksums.txt` | SHA256 checksums for verification | <1 KB |

### Version Verification

After release, users can verify the version:

```bash
./bundle-inspector version
# Output: Bundle Inspector 0.2.0
```

### Best Practices

1. **Test before release**: Always run `go test ./...` locally
2. **Clean git state**: Ensure no uncommitted changes
3. **Meaningful versions**: Follow semantic versioning (MAJOR.MINOR.PATCH)
4. **Descriptive tags**: Include brief release notes in tag annotation
5. **Verify assets**: Check that all binaries uploaded successfully
6. **Test binaries**: Download and test at least one binary per platform

## Future Enhancements (Not Yet Implemented)

- Compare command (delta between two artifacts)
- Baseline tracking (compare against saved baseline)
- Size regression prevention (built-in CI check)
- Android manifest parsing
- iOS entitlements analysis
- Detailed framework version reporting

## Debugging Tips

### Enable Verbose Output
```bash
# Set BITRISE environment variables locally for testing
export BITRISE_IPA_PATH=/path/to/test.ipa
export BITRISE_DEPLOY_DIR=/tmp/deploy
./bundle-inspector analyze
```

### Test Auto-Detection
```bash
# Test with environment variables
BITRISE_IPA_PATH=test.ipa ./bundle-inspector analyze

# Test without (should auto-detect)
./bundle-inspector analyze
```

### Check Bitrise Export
```bash
export BITRISE_DEPLOY_DIR=/tmp/test-deploy
./bundle-inspector analyze test.ipa
ls -la /tmp/test-deploy  # Should see reports
```

### Profile Performance
```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

## Quick Reference: Key Functions

### Main Entry Point
- `cmd/bundle-inspector/main.go:runAnalyze()`: Main analysis flow

### Analysis Orchestration
- `internal/orchestrator/orchestrator.go:RunAnalysis()`: Coordinates full analysis

### Bitrise Integration
- `internal/bitrise/env.go:DetectBundlePath()`: Auto-detect artifact
- `internal/bitrise/env.go:ExportToDeployDir()`: Export reports
- `internal/bitrise/env.go:IsBitriseEnvironment()`: Check if in Bitrise

### Output Generation
- `internal/report/text.go:Format()`: Text output
- `internal/report/json.go:Format()`: JSON output
- `internal/report/markdown.go:Format()`: Markdown output
- `internal/report/html.go:Format()`: HTML output

### Detection
- `internal/detector/duplicate.go:DetectDuplicates()`: Find duplicate files
- `internal/detector/bloat.go:DetectLargeFiles()`: Find large files

## Questions to Ask When Modifying

1. **Does this affect Bitrise integration?** Check `internal/bitrise/`
2. **Does this change the report structure?** Update `pkg/types/` and all formatters
3. **Does this add a new flag?** Update `main.go`, README.md, and QUICKSTART.md
4. **Does this add a new output format?** Add formatter, update CLI, update docs
5. **Does this change auto-detection?** Update `bitrise/env.go` and document priority
6. **Does this affect performance?** Benchmark with large artifacts
7. **Does this need tests?** Yes, always add tests
8. **Does this need documentation?** Update README.md and/or QUICKSTART.md

## Context for AI Assistants

When working on this codebase:

1. **Prioritize Bitrise integration**: This is a Bitrise plugin first, standalone CLI second
2. **Preserve auto-detection**: Don't break the auto-detection flow
3. **Maintain all 4 output formats**: Changes to report structure affect all formatters
4. **Test with real artifacts**: Don't rely on mocks for analysis features
5. **Update documentation**: README.md for users, QUICKSTART.md for developers
6. **Follow Go conventions**: Error wrapping, defer cleanup, idiomatic Go
7. **Consider performance**: Large artifacts (100MB+) are common
8. **Maintain backward compatibility**: Existing workflows depend on this tool

## External Dependencies

### Key Go Modules
- `github.com/spf13/cobra`: CLI framework
- `github.com/blacktop/lzfse-cgo`: LZFSE decompression (iOS)
- Standard library: `archive/zip`, `encoding/json`, `crypto/sha256`, etc.

### No External Runtime Dependencies
- Self-contained binary
- No Docker, Python, or other runtimes needed
- Works on macOS (Intel/ARM) and Linux

## Support & Resources

- **GitHub Issues**: Bug reports and feature requests
- **README.md**: User documentation
- **QUICKSTART.md**: Developer documentation
- **Source Code**: Well-commented, read the code for details
