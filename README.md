# Bitrise Bundle Inspector Plugin

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)]() [![License](https://img.shields.io/badge/license-Bitrise-blue)]()

A Bitrise CLI plugin that analyzes mobile artifacts (iOS and Android) for size optimization opportunities.

## Features

- **iOS & Android Support** - Analyze `.ipa`, `.app`, `.xcarchive` (iOS) and `.apk`, `.aab` (Android)
- **Duplicate Detection** - Find identical files with SHA-256 hashing
- **Optimization Recommendations** - Actionable suggestions with severity levels
- **4 Output Formats** - Text, JSON, Markdown, HTML with interactive visualizations
- **Auto-Detection** - Automatically detects artifact paths from Bitrise environment
- **Automatic Export** - Reports exported to Bitrise deploy directory for easy access
- **iOS Advanced Analysis** - Mach-O binary parsing, framework dependencies, Assets.car analysis

## Quick Start 🚀

### Installation

Choose one of the following methods:

#### 1. As Bitrise Plugin (Recommended)

```bash
bitrise plugin install https://github.com/bitrise-io/bitrise-plugins-bundle-inspector.git
```

Verify installation:
```bash
bitrise :bundle-inspector version
```

#### 2. Build from Source

```bash
git clone https://github.com/bitrise-io/bitrise-plugins-bundle-inspector.git
cd bitrise-plugins-bundle-inspector
go build -o bundle-inspector ./cmd/bundle-inspector
```

#### 3. Download Binary Release

Download pre-built binaries for macOS/Linux from [GitHub Releases](https://github.com/bitrise-io/bitrise-plugins-bundle-inspector/releases).

### Your First Analysis

Add this minimal workflow step to analyze your app:

```yaml
workflows:
  primary:
    steps:
    - xcode-archive@4:
    - script:
        title: Analyze Bundle Size
        inputs:
        - content: |
            #!/bin/bash
            set -ex
            bitrise :bundle-inspector analyze
```

That's it! The tool will:
- Auto-detect your IPA/APK/AAB from Bitrise environment variables
- Generate comprehensive analysis reports
- Export reports to `$BITRISE_DEPLOY_DIR` for download

**Expected outputs** in Build Artifacts:
- `bundle-analysis.json` - Machine-readable data
- `bundle-report.txt` - Human-readable report
- `bundle-analysis.md` - Markdown summary

## Usage 📊

### 1. Bitrise CI/CD Integration (Primary Use Case)

#### Basic Analysis with Auto-Detection

```yaml
workflows:
  primary:
    steps:
    - xcode-archive@4:
        # Generates BITRISE_IPA_PATH
    - script:
        title: Analyze Bundle Size
        inputs:
        - content: |
            #!/bin/bash
            set -ex
            bitrise :bundle-inspector analyze
            # Auto-detects from BITRISE_IPA_PATH
            # Exports to $BITRISE_DEPLOY_DIR
```

#### Size Limit Enforcement

Fail the build if bundle size exceeds a threshold:

```yaml
workflows:
  primary:
    steps:
    - xcode-archive@4:
    - script:
        title: Check Bundle Size
        inputs:
        - content: |
            #!/bin/bash
            set -ex

            # Run analysis
            bitrise :bundle-inspector analyze -o json -f analysis.json

            # Extract size (in bytes)
            SIZE=$(jq '.artifact_info.size' analysis.json)
            SIZE_MB=$((SIZE / 1024 / 1024))

            echo "Bundle size: ${SIZE_MB}MB"

            # Fail if over 50MB
            if [ $SIZE -gt 52428800 ]; then
              echo "❌ Bundle too large: ${SIZE_MB}MB exceeds 50MB limit"
              exit 1
            fi

            echo "✅ Bundle size OK"
```

#### Post Results to GitHub PR

Comment bundle analysis on pull requests:

```yaml
workflows:
  primary:
    steps:
    - xcode-archive@4:
    - script:
        title: Analyze and Comment on PR
        inputs:
        - content: |
            #!/bin/bash
            set -ex

            # Generate markdown report
            bitrise :bundle-inspector analyze -o markdown -f report.md

            # Post to GitHub PR (requires gh CLI)
            if [ -n "$PR_NUMBER" ]; then
              gh pr comment $PR_NUMBER --body-file report.md
            fi
```

#### Size Tracking Over Time

Track bundle size history across builds:

```yaml
workflows:
  primary:
    steps:
    - xcode-archive@4:
    - script:
        title: Track Bundle Size
        inputs:
        - content: |
            #!/bin/bash
            set -ex

            # Analyze bundle
            bitrise :bundle-inspector analyze -o json -f analysis.json

            # Extract key metrics
            SIZE=$(jq '.artifact_info.size' analysis.json)
            SAVINGS=$(jq '.total_savings' analysis.json)
            BUILD_NUM=$BITRISE_BUILD_NUMBER

            # Append to tracking file (store in artifact storage)
            echo "$BUILD_NUM,$SIZE,$SAVINGS" >> size-history.csv

            # Upload to your tracking service
            # curl -X POST https://your-tracking-service.com/metrics \
            #   -d "build=$BUILD_NUM&size=$SIZE&savings=$SAVINGS"
```

#### Multi-Flavor Analysis

Compare different build variants:

```yaml
workflows:
  analyze-all-flavors:
    steps:
    - android-build@1:
        inputs:
        - variant: debug
        - output_file: app-debug.apk
    - script:
        title: Analyze Debug Build
        inputs:
        - content: |
            bitrise :bundle-inspector analyze app-debug.apk -f debug-report.json -o json

    - android-build@1:
        inputs:
        - variant: release
        - output_file: app-release.apk
    - script:
        title: Analyze Release Build
        inputs:
        - content: |
            bitrise :bundle-inspector analyze app-release.apk -f release-report.json -o json

    - script:
        title: Compare Builds
        inputs:
        - content: |
            #!/bin/bash
            DEBUG_SIZE=$(jq '.artifact_info.size' debug-report.json)
            RELEASE_SIZE=$(jq '.artifact_info.size' release-report.json)
            DIFF=$((RELEASE_SIZE - DEBUG_SIZE))

            echo "Debug: $((DEBUG_SIZE / 1024 / 1024))MB"
            echo "Release: $((RELEASE_SIZE / 1024 / 1024))MB"
            echo "Difference: $((DIFF / 1024 / 1024))MB"
```

### 2. Local Plugin Usage

When installed as a Bitrise plugin, use the `:` prefix:

```bash
# Auto-detect from Bitrise environment variables
bitrise :bundle-inspector analyze

# Explicit path
bitrise :bundle-inspector analyze path/to/app.ipa

# Single output format
bitrise :bundle-inspector analyze -o json -f report.json

# Multiple output formats (runs analysis once, generates all formats)
bitrise :bundle-inspector analyze -o json,markdown,html
# Creates: bundle-analysis-<artifact>.json, .md, .html

# Multiple formats with custom filenames
bitrise :bundle-inspector analyze -o json,html -f report.json,report.html

# Disable auto-detection
bitrise :bundle-inspector analyze --no-auto-detect path/to/app.ipa
```

### 3. Standalone CLI

Use the binary directly without Bitrise plugin:

```bash
# Basic analysis (text format)
./bundle-inspector analyze app.ipa

# Single format with custom file
./bundle-inspector analyze app.ipa -o json -f report.json

# Multiple formats at once (efficient - runs analysis only once)
./bundle-inspector analyze app.ipa -o json,markdown,html
# Creates: bundle-analysis-app.json, bundle-analysis-app.md, bundle-analysis-app.html

# All 4 formats
./bundle-inspector analyze app.apk -o text,json,markdown,html

# Custom filenames for multiple formats
./bundle-inspector analyze app.aab -o json,html -f data.json,interactive.html
```

## Configuration ⚙️

### Environment Variables

The plugin automatically detects artifacts from Bitrise environment variables:

| Variable | Description | Priority | Example |
|----------|-------------|----------|---------|
| `BITRISE_IPA_PATH` | iOS IPA path | 1st | `/tmp/MyApp.ipa` |
| `BITRISE_AAB_PATH` | Android AAB path | 2nd | `/tmp/app.aab` |
| `BITRISE_APK_PATH` | Android APK path | 3rd | `/tmp/app.apk` |
| `BITRISE_DEPLOY_DIR` | Output directory | - | `/tmp/deploy` |
| `BITRISE_BUILD_NUMBER` | Build number | - | `123` |
| `GIT_CLONE_COMMIT_HASH` | Git commit hash | - | `abc123...` |

#### Auto-Detection Behavior

**Priority order:** IPA → AAB → APK

The tool checks each variable in order and uses the first valid file found:

1. Checks `BITRISE_IPA_PATH` - if file exists, uses it
2. Checks `BITRISE_AAB_PATH` - if file exists, uses it
3. Checks `BITRISE_APK_PATH` - if file exists, uses it
4. If none found, returns error

**Override auto-detection:**

```bash
# Provide explicit path (takes precedence)
bitrise :bundle-inspector analyze /custom/path/to/app.ipa

# Disable auto-detection completely
bitrise :bundle-inspector analyze --no-auto-detect /path/to/app.ipa
```

#### Deploy Directory Integration

When `BITRISE_DEPLOY_DIR` is set, reports are automatically exported:

- `bundle-analysis.json` - JSON format
- `bundle-report.txt` - Text format
- `bundle-analysis.md` - Markdown format

These files appear in **Build Artifacts** for easy download and viewing.

### Command Flags

Complete reference of available flags:

```bash
bundle-inspector analyze [file-path] [flags]

Flags:
  -o, --output string         Output format(s) - comma-separated for multiple
                              Valid formats: text, json, markdown, html (default "text")
  -f, --output-file string    Output filename(s) - comma-separated when using multiple formats
                              (default: auto-generated as bundle-analysis-<artifact>.<ext>)
      --include-duplicates    Enable duplicate file detection (default true)
      --no-auto-detect        Disable auto-detection from Bitrise environment
  -h, --help                  Help for analyze
```

#### Flag Examples

```bash
# Single output format
bitrise :bundle-inspector analyze -o json

# Multiple output formats (comma-separated)
bitrise :bundle-inspector analyze -o json,markdown,html

# All 4 formats at once
bitrise :bundle-inspector analyze -o text,json,markdown,html

# Custom filename (single format)
bitrise :bundle-inspector analyze -o json -f my-analysis.json

# Custom filenames (multiple formats, must match count)
bitrise :bundle-inspector analyze -o json,html -f data.json,report.html

# Disable duplicate detection (faster)
bitrise :bundle-inspector analyze --include-duplicates=false

# Explicit path, no auto-detect
bitrise :bundle-inspector analyze --no-auto-detect /path/to/app.ipa
```

#### Default Output Filenames

When `-f` is not specified, filenames are auto-generated:

- **Text:** `bundle-analysis-<artifact>.txt`
- **JSON:** `bundle-analysis-<artifact>.json`
- **Markdown:** `bundle-analysis-<artifact>.md`
- **HTML:** `bundle-analysis-<artifact>.html`

Example: Analyzing `MyApp.ipa` creates `bundle-analysis-MyApp.txt`

#### Multi-Format Output Benefits

Generate multiple formats in a single analysis run:

```bash
# Efficient: Analysis runs once, generates 3 formats
bitrise :bundle-inspector analyze app.ipa -o json,markdown,html
```

**Advantages:**
- ✅ **Faster**: Analysis runs only once (important for large artifacts)
- ✅ **Consistent**: All formats contain identical data from same analysis
- ✅ **Convenient**: Get multiple report types for different audiences
  - JSON for automation/scripting
  - Markdown for PR comments
  - HTML for stakeholder presentations

## Output Formats

### 1. Text (Human-Readable)

Best for console output and quick inspection:

```bash
bitrise :bundle-inspector analyze app.ipa -o text
```

**Use cases:**
- Quick local inspection
- CI/CD logs
- Debugging

**Sample output:**
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

Optimization Opportunities:

  High Priority:
    • Remove duplicate files
      Found 2 identical files (1.2 MB each)
      Potential savings: 1.2 MB
```

### 2. JSON (CI/CD Automation)

Best for parsing, automation, and programmatic access:

```bash
bitrise :bundle-inspector analyze app.ipa -o json -f report.json
```

**Use cases:**
- Size limit checks
- Automated comparisons
- Data extraction for dashboards
- Integration with other tools

**Sample output:**
```json
{
  "artifact_info": {
    "path": "MyApp.ipa",
    "type": "ipa",
    "size": 47448064,
    "uncompressed_size": 55353344,
    "analyzed_at": "2026-02-02T10:30:00Z"
  },
  "size_breakdown": {
    "executable": 16040960,
    "frameworks": 23168000,
    "assets": 8806400,
    "resources": 5452800,
    "by_category": {
      "executable": 16040960,
      "frameworks": 23168000,
      "assets": 8806400
    },
    "by_extension": {
      ".dylib": 23168000,
      ".car": 6500000
    }
  },
  "duplicates": [
    {
      "hash": "abc123...",
      "paths": ["Assets/image.png", "Resources/image.png"],
      "size": 1258291,
      "count": 2,
      "wasted_space": 1258291
    }
  ],
  "optimizations": [
    {
      "severity": "high",
      "title": "Remove duplicate files",
      "description": "Found 2 identical files",
      "potential_savings": 1258291,
      "action": "Keep only one copy and deduplicate references"
    }
  ],
  "total_savings": 1258291
}
```

**Parsing examples:**

```bash
# Extract total size
jq '.artifact_info.size' report.json

# Get high-priority optimizations
jq '.optimizations[] | select(.severity == "high")' report.json

# Calculate size in MB
jq '.artifact_info.size / 1024 / 1024' report.json
```

### 3. Markdown (PR Comments & Documentation)

Best for pull request comments, Slack messages, and documentation:

```bash
bitrise :bundle-inspector analyze app.ipa -o markdown -f report.md
```

**Use cases:**
- GitHub/GitLab PR comments
- Slack notifications
- Documentation generation
- Team reports

**Sample output:**
```markdown
# Bundle Analysis Report

## Artifact Information
- **Type:** ipa
- **Path:** MyApp.ipa
- **Size:** 45.2 MB (compressed)
- **Uncompressed:** 52.8 MB
- **Compression:** 85.6%

## Size Breakdown
| Category | Size | Percentage |
|----------|------|------------|
| Executable | 15.3 MB | 29.0% |
| Frameworks | 22.1 MB | 41.9% |
| Assets | 8.4 MB | 15.9% |

## Optimization Opportunities

### High Priority
- **Remove duplicate files**
  - Found 2 identical files (1.2 MB each)
  - Potential savings: 1.2 MB
  - Action: Keep only one copy

**Total Potential Savings:** 1.2 MB (2.3%)
```

**Usage with GitHub CLI:**

```bash
# Post to PR
gh pr comment $PR_NUMBER --body-file report.md

# Add to PR description
gh pr create --title "Update" --body-file report.md
```

### 4. HTML (Interactive Visualizations)

Best for stakeholder reports and interactive exploration:

```bash
bitrise :bundle-inspector analyze app.ipa -o html -f report.html
```

**Use cases:**
- Interactive visualizations
- Stakeholder presentations
- Detailed exploration
- Tree map views

**Features:**
- Interactive treemap visualization
- Sortable tables
- Collapsible sections
- Size percentage bars
- Mobile-responsive design

**Access report:**
1. Download from Build Artifacts
2. Open in browser
3. Explore interactively

### Output Destinations

#### Default Behavior

```bash
# Without -f flag, creates file in current directory
bitrise :bundle-inspector analyze app.ipa -o json
# Creates: bundle-analysis-app.json

# With -f flag, uses specified path
bitrise :bundle-inspector analyze app.ipa -o json -f /custom/path/report.json
# Creates: /custom/path/report.json
```

#### Bitrise Deploy Directory (Automatic)

When running in Bitrise CI:
- Reports automatically exported to `$BITRISE_DEPLOY_DIR`
- Available in **Build Artifacts** tab
- Multiple formats exported simultaneously:
  - `bundle-analysis.json`
  - `bundle-report.txt`
  - `bundle-analysis.md`

## Advanced Use Cases

### Size Regression Prevention

Prevent accidental size increases in CI:

```bash
#!/bin/bash
set -e

# Analyze current build
bitrise :bundle-inspector analyze -o json -f current.json

# Compare with baseline
CURRENT_SIZE=$(jq '.artifact_info.size' current.json)
BASELINE_SIZE=$(cat baseline-size.txt)
INCREASE=$((CURRENT_SIZE - BASELINE_SIZE))
PERCENT=$((INCREASE * 100 / BASELINE_SIZE))

if [ $PERCENT -gt 5 ]; then
  echo "❌ Size increased by ${PERCENT}%"
  exit 1
fi

# Update baseline if passed
echo $CURRENT_SIZE > baseline-size.txt
```

### PR Comments with Size Deltas

Show size changes in pull requests:

```bash
#!/bin/bash
set -e

# Analyze current PR build
bitrise :bundle-inspector analyze -o json -f pr.json

# Fetch main branch build (from artifact storage)
curl -o main.json https://your-storage.com/main-branch-analysis.json

# Calculate delta
PR_SIZE=$(jq '.artifact_info.size' pr.json)
MAIN_SIZE=$(jq '.artifact_info.size' main.json)
DELTA=$((PR_SIZE - MAIN_SIZE))
DELTA_MB=$((DELTA / 1024 / 1024))

# Format comment
cat > comment.md <<EOF
## Bundle Size Analysis

| Metric | Main | PR | Delta |
|--------|------|----|----|
| Size | $((MAIN_SIZE / 1024 / 1024))MB | $((PR_SIZE / 1024 / 1024))MB | ${DELTA_MB}MB |

$(if [ $DELTA -gt 0 ]; then echo "⚠️ Size increased"; else echo "✅ Size decreased"; fi)
EOF

gh pr comment $PR_NUMBER --body-file comment.md
```

### Historical Tracking Dashboard

Build a size tracking dashboard:

```bash
#!/bin/bash
set -e

# Analyze and extract metrics
bitrise :bundle-inspector analyze -o json -f analysis.json

SIZE=$(jq '.artifact_info.size' analysis.json)
SAVINGS=$(jq '.total_savings' analysis.json)
BUILD=$BITRISE_BUILD_NUMBER
COMMIT=$GIT_CLONE_COMMIT_HASH

# Send to time-series database
curl -X POST https://your-metrics.com/api/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
    "build": "'$BUILD'",
    "commit": "'$COMMIT'",
    "size": '$SIZE',
    "potential_savings": '$SAVINGS'
  }'
```

### Multi-Variant Comparison

Compare debug vs release builds:

```bash
#!/bin/bash
set -e

# Analyze both variants
bitrise :bundle-inspector analyze debug.apk -o json -f debug.json
bitrise :bundle-inspector analyze release.apk -o json -f release.json

# Extract sizes
DEBUG_SIZE=$(jq '.artifact_info.size' debug.json)
RELEASE_SIZE=$(jq '.artifact_info.size' release.json)
REDUCTION=$((DEBUG_SIZE - RELEASE_SIZE))
PERCENT=$((REDUCTION * 100 / DEBUG_SIZE))

echo "Debug: $((DEBUG_SIZE / 1024 / 1024))MB"
echo "Release: $((RELEASE_SIZE / 1024 / 1024))MB"
echo "Reduction: ${PERCENT}% ($((REDUCTION / 1024 / 1024))MB)"
```

## Supported Artifacts

### iOS
- **.ipa** - iOS App Package (full support)
- **.app** - iOS App Bundle (full support)
- **.xcarchive** - Xcode Archive (full support)

### Android
- **.apk** - Android Package (full support)
- **.aab** - Android App Bundle (full support)

## Understanding the Output

### Report Sections

#### 1. Artifact Information
- **Type**: File format (ipa, apk, aab)
- **Path**: File location
- **Compressed Size**: Archive size on disk
- **Uncompressed Size**: Extracted content size
- **Compression Ratio**: How much compression reduces size

#### 2. Size Breakdown
Categorizes files by type:
- **Executable**: Main app binary
- **Frameworks**: Dynamic libraries (.dylib, .framework)
- **Libraries**: Native libraries (.so files for Android)
- **Assets**: Images, Assets.car (iOS), drawable resources (Android)
- **Resources**: Other resources (strings, configs, etc.)
- **DEX**: Dalvik bytecode (Android only)
- **Other**: Uncategorized files

#### 3. Top Largest Files
Lists the 10 largest individual files with:
- Full path within archive
- Size in MB
- Percentage of total

#### 4. Duplicate Files
Groups identical files (matched by SHA-256):
- Hash of file content
- All paths where file appears
- Size of each copy
- **Wasted Space**: Total redundant storage

#### 5. Optimization Opportunities
Actionable recommendations grouped by severity:

**High Priority**: Immediate attention needed
- Duplicate files
- Very large files (>5MB)
- Unused frameworks

**Medium Priority**: Good to address
- Large files (2-5MB)
- Uncompressed assets
- Debug symbols in release builds

**Low Priority**: Nice to have
- Moderate files (1-2MB)
- Minor optimizations

Each recommendation includes:
- **Title**: What to do
- **Description**: Why it matters
- **Potential Savings**: Size reduction estimate
- **Action**: How to fix it

#### 6. Total Potential Savings
Sum of all optimization opportunities with percentage of total size.

### Optimization Recommendations Guide

#### Remove Duplicate Files
**What it means**: Multiple identical copies of the same file exist

**How to fix**:
- Use asset catalogs (iOS)
- Enable resource shrinking (Android)
- Check for accidental duplicates in project

**Example**: Same image in both `Assets/` and `Resources/`

#### Compress Large Images
**What it means**: Uncompressed or poorly compressed images

**How to fix**:
- Use asset catalogs with automatic optimization (iOS)
- Use WebP format (Android)
- Run image optimization tools (ImageOptim, TinyPNG)

#### Remove Unused Frameworks
**What it means**: Frameworks included but not linked by app

**How to fix**:
- Review embedded frameworks
- Remove from "Embed Frameworks" build phase
- Clean up CocoaPods/Carthage dependencies

#### Strip Debug Symbols
**What it means**: Debug information included in release build

**How to fix**:
- Enable "Strip Debug Symbols" in Release configuration
- Set `COPY_PHASE_STRIP = YES`
- Use `strip` command manually

## iOS Advanced Analysis

The plugin includes deep iOS binary inspection capabilities:

- **Mach-O Binary Parsing**: Architecture detection (arm64, x86_64), binary type, code/data sizes
- **Framework Dependency Analysis**: Automatic discovery, dependency graphs, unused framework detection
- **Assets.car Parsing**: Asset extraction, type/scale categorization (@1x, @2x, @3x)
- **LZFSE Compression Support**: Automatic decompression of modern iOS IPAs

For detailed information, see [iOS Advanced Analysis Documentation](docs/ios-advanced-analysis.md).

## Development 🔧

### For Developers

See the [Developer Guide](QUICKSTART.md) for:
- Building from source
- Local development setup
- Testing with test artifacts
- Project structure
- Contributing guidelines

### Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## Troubleshooting

### "no bundle found in Bitrise environment variables"

**Cause**: No artifact path found in `BITRISE_IPA_PATH`, `BITRISE_AAB_PATH`, or `BITRISE_APK_PATH`

**Solutions**:
- Ensure you've run a build step that generates these variables (e.g., `xcode-archive`, `android-build`)
- Check step order - analysis must come after build
- Provide explicit path: `bitrise :bundle-inspector analyze /path/to/app.ipa`

### "BITRISE_DEPLOY_DIR not set"

**Cause**: Deploy directory not configured (rare in Bitrise)

**Solutions**:
- This is a warning, not an error - analysis still runs
- Reports saved to current directory instead
- Check Bitrise build settings

### "failed to open artifact"

**Cause**: File doesn't exist or insufficient permissions

**Solutions**:
- Verify file path: `ls -la /path/to/artifact`
- Check file was created by previous step
- Ensure file permissions allow reading

### "artifact not found"

**Cause**: Specified file path doesn't exist

**Solutions**:
- Double-check file path
- Use auto-detection instead: `bitrise :bundle-inspector analyze`
- Verify artifact was built successfully

### Slow Analysis Performance

**Cause**: Duplicate detection is computationally expensive

**Solutions**:
- Disable duplicate detection: `--include-duplicates=false`
- Normal for large apps (>100MB)
- Use faster output format (text instead of HTML)

### "Plugin not found" / "command not found"

**Cause**: Plugin not installed or not in PATH

**Solutions**:
- Install plugin: `bitrise plugin install https://github.com/bitrise-io/bitrise-plugins-bundle-inspector.git`
- Use direct binary: `./bundle-inspector` instead of `bitrise :bundle-inspector`
- Check PATH: `echo $PATH`

## License

Copyright © Bitrise