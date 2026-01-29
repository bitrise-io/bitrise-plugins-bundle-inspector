# Bundle Inspector - Quick Start Guide

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/bitrise-io/bitrise-plugins-bundle-inspector.git
cd bitrise-plugins-bundle-inspector

# Build the binary
go build -o bundle-inspector ./cmd/bundle-inspector

# Optional: Install to PATH
go install ./cmd/bundle-inspector
```

### Verify Installation

```bash
./bundle-inspector version
```

## Basic Usage

### Analyze an iOS App

```bash
# Analyze IPA file
./bundle-inspector analyze MyApp.ipa

# Save report to file
./bundle-inspector analyze MyApp.ipa -f report.txt
```

### Analyze an Android App

```bash
# Analyze APK
./bundle-inspector analyze MyApp.apk

# Analyze AAB (App Bundle)
./bundle-inspector analyze MyApp.aab
```

### JSON Output (for CI/CD)

```bash
# Generate JSON report
./bundle-inspector analyze MyApp.ipa -o json -f report.json

# Pretty-print to stdout
./bundle-inspector analyze MyApp.ipa -o json
```

## Advanced Options

### Custom Threshold for Large Files

```bash
# Warn about files larger than 500KB
./bundle-inspector analyze MyApp.ipa --threshold 500KB

# Warn about files larger than 2MB
./bundle-inspector analyze MyApp.ipa --threshold 2MB
```

### Disable Duplicate Detection

```bash
# Skip duplicate detection (faster)
./bundle-inspector analyze MyApp.ipa --include-duplicates=false
```

## Understanding the Output

### Text Output Sections

1. **Artifact Information**
   - File path, type, and sizes
   - Compression ratio for archives

2. **Size Breakdown**
   - Categories: Executable, Frameworks, Libraries, Assets, Resources, DEX, Other
   - Percentages relative to total size

3. **Top Largest Files**
   - Shows the 10 largest files with sizes and percentages

4. **Duplicate Files** (if detected)
   - Groups of identical files
   - Wasted space calculation

5. **Optimization Opportunities**
   - Grouped by severity (High, Medium, Low)
   - Estimated savings for each recommendation
   - Actionable suggestions

6. **Total Potential Savings**
   - Sum of all optimization opportunities

### JSON Output Structure

```json
{
  "artifact_info": {
    "path": "...",
    "type": "ipa|apk|aab",
    "size": 12345678,
    "uncompressed_size": 23456789,
    "analyzed_at": "2026-01-29T..."
  },
  "size_breakdown": {
    "executable": 1000000,
    "frameworks": 5000000,
    "assets": 3000000,
    "by_category": {...},
    "by_extension": {...}
  },
  "file_tree": [...],
  "duplicates": [...],
  "optimizations": [...],
  "largest_files": [...],
  "total_savings": 1000000
}
```

## CI/CD Integration Examples

### GitHub Actions

```yaml
- name: Analyze Bundle Size
  run: |
    ./bundle-inspector analyze app.ipa -o json -f report.json

    # Extract total size
    SIZE=$(jq '.artifact_info.size' report.json)
    echo "Bundle size: $SIZE bytes"

    # Fail if over 50MB
    if [ $SIZE -gt 52428800 ]; then
      echo "Bundle too large!"
      exit 1
    fi
```

### Extract Optimization Opportunities

```bash
# Get high-priority optimizations
./bundle-inspector analyze app.ipa -o json | \
  jq '.optimizations[] | select(.severity == "high")'

# Get total potential savings
./bundle-inspector analyze app.ipa -o json | \
  jq '.total_savings'
```

### Compare Bundle Sizes

```bash
# Analyze current build
./bundle-inspector analyze current.ipa -o json -f current.json

# Compare with previous
CURRENT_SIZE=$(jq '.artifact_info.size' current.json)
PREVIOUS_SIZE=$(jq '.artifact_info.size' previous.json)
DIFF=$((CURRENT_SIZE - PREVIOUS_SIZE))

echo "Size change: $DIFF bytes"
```

## Troubleshooting

### "artifact not found"
- Verify the file path is correct
- Check file permissions

### "no .app bundle found in IPA"
- IPA file may be corrupted
- Ensure it's a valid iOS IPA archive

### "failed to open APK"
- APK/AAB file may be corrupted
- Verify it's a valid ZIP archive

### Slow Analysis
- Large artifacts take longer to analyze
- Duplicate detection adds overhead
- Try `--include-duplicates=false` for faster analysis

## Tips

1. **Regular Monitoring**: Run analysis on every build to track size trends
2. **Set Thresholds**: Use `--threshold` to catch large file additions early
3. **Focus on High-Priority**: Address high-severity optimizations first
4. **Automate**: Integrate JSON output into CI/CD for automated checks
5. **Compare Builds**: Track size changes between releases

## Getting Help

```bash
# Show help
./bundle-inspector --help

# Command-specific help
./bundle-inspector analyze --help
```

## Next Steps

- Read the full [README.md](README.md) for detailed features
- Check [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for technical details
- Explore the source code in `internal/` and `pkg/` directories
- Contribute improvements via pull requests
