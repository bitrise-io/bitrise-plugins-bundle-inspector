# Feature: Multi-Format Output Support

## Overview

Added support for generating multiple output formats in a single analysis run. Users can now specify comma-separated formats to generate all reports efficiently.

## Implementation

### Changes Made

1. **cmd/bundle-inspector/main.go**:
   - Changed `outputFormat` → `outputFormats` (comma-separated string)
   - Changed `outputFile` → `outputFiles` (comma-separated string)
   - Added `parseFormats()` - Parse and validate comma-separated formats
   - Added `parseOutputFiles()` - Parse comma-separated filenames
   - Added `getFileExtension()` - Get extension for each format
   - Updated `determineOutputFiles()` - Generate filenames for all formats
   - Updated `writeReport()` - Accept format parameter instead of global
   - Updated `runAnalyze()` - Loop through formats and generate all reports

2. **README.md**:
   - Updated "Local Plugin Usage" section with multi-format examples
   - Updated "Standalone CLI" section with multi-format examples
   - Updated "Command Flags" section with new flag descriptions
   - Added "Multi-Format Output Benefits" section

3. **QUICKSTART.md** (Developer Guide):
   - Updated test artifact examples to use multi-format

4. **claude.md**:
   - Updated CLI flags documentation
   - Updated usage patterns with multi-format examples
   - Added "Multi-Format Output" section in implementation details

## Usage

### Basic Multi-Format

```bash
# Generate JSON, Markdown, and HTML reports
./bundle-inspector analyze app.ipa -o json,markdown,html

# Output:
#   bundle-analysis-app.json
#   bundle-analysis-app.md
#   bundle-analysis-app.html
```

### All 4 Formats

```bash
./bundle-inspector analyze app.ipa -o text,json,markdown,html
```

### Custom Filenames

```bash
# Number of filenames must match number of formats
./bundle-inspector analyze app.ipa -o json,html -f data.json,report.html
```

### With Bitrise Plugin

```bash
bitrise :bundle-inspector analyze -o json,markdown,html
```

## Benefits

1. **Performance**: Analysis runs only once (important for large artifacts >100MB)
2. **Consistency**: All formats contain identical data from same analysis
3. **Convenience**: One command generates reports for multiple audiences:
   - JSON for automation/CI checks
   - Markdown for PR comments
   - HTML for stakeholder presentations
   - Text for console output

## Validation

### Format Validation

```bash
$ ./bundle-inspector analyze app.ipa -o json,xml,html
Error: unsupported output format: xml (valid formats: text, json, markdown, html)
```

### Filename Count Validation

```bash
$ ./bundle-inspector analyze app.ipa -o json,html,markdown -f only-one.json
Error: number of output files (1) must match number of formats (3)
```

### Duplicate Format Handling

Duplicate formats in the list are automatically deduplicated:

```bash
./bundle-inspector analyze app.ipa -o json,json,html
# Only generates: json, html (duplicate json ignored)
```

## Testing

### Test Cases

1. ✅ Single format (backward compatible)
2. ✅ Multiple formats with auto-generated filenames
3. ✅ Multiple formats with custom filenames
4. ✅ All 4 formats at once
5. ✅ Invalid format error handling
6. ✅ Mismatched filename count error handling
7. ✅ Duplicate format deduplication

### Test Example

```bash
# Test with real artifact
./bundle-inspector analyze test-artifacts/android/2048-game-2048.apk -o json,markdown,html

# Output:
Analyzing test-artifacts/android/2048-game-2048.apk...
Detecting duplicates and additional optimizations...

Generating reports:
  ✓ JSON: bundle-analysis-2048-game-2048.json
  ✓ MARKDOWN: bundle-analysis-2048-game-2048.md
  ✓ HTML: bundle-analysis-2048-game-2048.html
```

## Backward Compatibility

✅ **Fully backward compatible**

Existing single-format usage continues to work:

```bash
# Old usage still works
./bundle-inspector analyze app.ipa -o json
./bundle-inspector analyze app.ipa -o json -f report.json
```

## Code Changes Summary

- **Files Modified**: 4 (main.go, README.md, QUICKSTART.md, claude.md)
- **Lines Added**: ~150
- **Functions Added**: 3 (parseFormats, parseOutputFiles, getFileExtension)
- **Functions Modified**: 3 (determineOutputFiles, writeReport, runAnalyze)
- **Breaking Changes**: None

## Future Enhancements

Potential improvements:

1. Add `--all-formats` flag as shorthand for `-o text,json,markdown,html`
2. Add format-specific options (e.g., `--json-compact` for compact JSON)
3. Support output to directories: `-o json,html --output-dir ./reports/`
4. Progress indicator for multi-format generation on large artifacts

## Documentation

All documentation updated:
- ✅ README.md - User-facing examples
- ✅ QUICKSTART.md - Developer examples
- ✅ claude.md - AI assistant context
- ✅ Help text - `./bundle-inspector analyze --help`
