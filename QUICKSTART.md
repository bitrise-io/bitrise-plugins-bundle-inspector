# Developer Guide - Bundle Inspector

This guide covers development workflows, building from source, testing, and contributing to Bundle Inspector.

## Table of Contents

- [Building from Source](#building-from-source)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Output Formats](#output-formats)
- [Contributing](#contributing)

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git

### Clone and Build

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

### Build for Multiple Platforms

```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o bundle-inspector-darwin-amd64 ./cmd/bundle-inspector

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bundle-inspector-darwin-arm64 ./cmd/bundle-inspector

# Linux
GOOS=linux GOARCH=amd64 go build -o bundle-inspector-linux-amd64 ./cmd/bundle-inspector
```

## Development Setup

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/analyzer/...
go test ./internal/report/...

# Verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Formatting and Linting

```bash
# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Install golangci-lint
# macOS: brew install golangci-lint
# Linux: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Development Workflow

1. **Make changes** to source code
2. **Build**: `go build -o bundle-inspector ./cmd/bundle-inspector`
3. **Test**: Run unit tests for affected packages
4. **Test integration**: Run against test artifacts
5. **Format**: `go fmt ./...`
6. **Commit**: Follow conventional commit format

## Project Structure

```
bundle-inspector/
├── cmd/
│   └── bundle-inspector/          # CLI entry point
│       └── main.go                # Command-line interface, flags, orchestration
├── internal/
│   ├── analyzer/                  # Artifact analyzers
│   │   ├── ios/                   # iOS-specific analyzers
│   │   │   ├── ipa.go             # IPA file analyzer
│   │   │   ├── app.go             # .app bundle analyzer
│   │   │   ├── xcarchive.go       # XCArchive analyzer
│   │   │   ├── macho/             # Mach-O binary parser
│   │   │   ├── assetcar/          # Assets.car parser
│   │   │   └── framework/         # Framework dependency analyzer
│   │   └── android/               # Android-specific analyzers
│   │       ├── apk.go             # APK analyzer
│   │       └── aab.go             # AAB analyzer
│   ├── bitrise/                   # Bitrise integration
│   │   └── env.go                 # Environment variable handling
│   ├── detector/                  # Detection algorithms
│   │   ├── duplicate.go           # Duplicate file detection
│   │   └── bloat.go               # Large file detection
│   ├── orchestrator/              # Analysis orchestration
│   │   └── orchestrator.go        # Coordinates analyzers and detectors
│   ├── report/                    # Output formatters
│   │   ├── text.go                # Human-readable text output
│   │   ├── json.go                # JSON output
│   │   ├── markdown.go            # Markdown output
│   │   └── html.go                # HTML with treemap visualizations
│   └── util/                      # Utility functions
│       ├── humanize.go            # Size formatting
│       └── zip.go                 # ZIP archive utilities
├── pkg/
│   └── types/                     # Public API types
│       └── types.go               # Report structures, data models
├── test-artifacts/                # Real mobile apps for testing
│   ├── ios/
│   │   ├── lightyear.ipa          # Real iOS app (81MB)
│   │   └── Wikipedia.app/         # Real .app bundle
│   └── android/
│       └── 2048-game-2048.apk     # Real Android app (11MB)
├── scripts/                       # Helper scripts
│   ├── analyze-all-test-artifacts.sh
│   └── run-integration-tests.sh
├── bitrise-plugin.yml             # Bitrise plugin configuration
├── go.mod                         # Go module dependencies
└── README.md                      # User-facing documentation
```

### Key Packages

#### `cmd/bundle-inspector`
CLI entry point using Cobra. Handles:
- Command-line parsing
- Flag management
- Output file handling
- Bitrise environment integration

#### `internal/analyzer`
Platform-specific analyzers:
- **iOS**: IPA, .app, XCArchive support with Mach-O parsing, framework analysis, Assets.car extraction
- **Android**: APK, AAB support with DEX enumeration, native library detection

#### `internal/detector`
Detection algorithms:
- **Duplicate detection**: SHA-256 hashing with parallel processing
- **Bloat detection**: Large file identification with configurable thresholds

#### `internal/orchestrator`
Coordinates analysis workflow:
1. Detect artifact type
2. Run appropriate analyzer
3. Run detectors (duplicates, bloat)
4. Generate optimization recommendations
5. Build report

#### `internal/report`
Output formatters:
- **Text**: Console-friendly format
- **JSON**: Machine-parseable format
- **Markdown**: GitHub/GitLab-friendly format
- **HTML**: Interactive treemap visualizations

#### `pkg/types`
Public API types and data structures:
- `Report`: Complete analysis report
- `ArtifactInfo`: Artifact metadata
- `SizeBreakdown`: Size categorization
- `Duplicate`: Duplicate file groups
- `Optimization`: Optimization recommendations

## Testing

### Test Artifacts

The project includes real mobile applications for comprehensive testing:

```bash
test-artifacts/
├── ios/
│   ├── lightyear.ipa          # 81MB iOS game with complex structure
│   └── Wikipedia.app/         # Real .app bundle with frameworks
└── android/
    └── 2048-game-2048.apk     # 11MB Android game
```

**Note**: Test artifacts are large binaries excluded from git. Store them locally in `test-artifacts/`.

### Running Against Test Artifacts

```bash
# Analyze iOS IPA
./bundle-inspector analyze test-artifacts/ios/lightyear.ipa

# Analyze iOS .app bundle
./bundle-inspector analyze test-artifacts/ios/Wikipedia.app

# Analyze Android APK
./bundle-inspector analyze test-artifacts/android/2048-game-2048.apk

# Generate all 4 format reports (efficient - runs analysis once)
./bundle-inspector analyze test-artifacts/ios/lightyear.ipa -o text,json,markdown,html

# Or individual formats
./bundle-inspector analyze test-artifacts/ios/lightyear.ipa -o json -f report.json
./bundle-inspector analyze test-artifacts/ios/lightyear.ipa -o html -f report.html
```

### Integration Test Script

```bash
# Run all analyses with reports
./scripts/analyze-all-test-artifacts.sh

# Run integration tests
./scripts/run-integration-tests.sh
```

### Unit Tests

Write unit tests for new functionality:

```go
package analyzer

import "testing"

func TestIPAAnalyzer(t *testing.T) {
    analyzer := NewIPAAnalyzer()
    report, err := analyzer.Analyze("test-artifacts/ios/test.ipa")
    if err != nil {
        t.Fatalf("analysis failed: %v", err)
    }

    if report.ArtifactInfo.Type != "ipa" {
        t.Errorf("expected type 'ipa', got '%s'", report.ArtifactInfo.Type)
    }
}
```

### Testing Best Practices

1. **Use real artifacts**: Test with actual IPAs/APKs, not mocked data
2. **Test all formats**: Verify text, JSON, markdown, and HTML outputs
3. **Test edge cases**: Empty archives, corrupted files, unusual structures
4. **Performance tests**: Measure analysis time for large artifacts
5. **Integration tests**: Test full workflow from CLI to output

## Output Formats

### Adding a New Output Format

To add a new output format (e.g., XML):

1. **Create formatter**: `internal/report/xml.go`

```go
package report

import (
    "encoding/xml"
    "io"
    "github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

type XMLFormatter struct{}

func NewXMLFormatter() *XMLFormatter {
    return &XMLFormatter{}
}

func (f *XMLFormatter) Format(w io.Writer, report *types.Report) error {
    encoder := xml.NewEncoder(w)
    encoder.Indent("", "  ")
    return encoder.Encode(report)
}
```

2. **Update CLI**: `cmd/bundle-inspector/main.go`

```go
case "xml":
    formatter := report.NewXMLFormatter()
    if err := formatter.Format(f, analysisReport); err != nil {
        return fmt.Errorf("failed to format output: %w", err)
    }
```

3. **Add tests**: `internal/report/xml_test.go`

4. **Update documentation**: Update README.md with new format

### Understanding Existing Formatters

#### Text Formatter (`internal/report/text.go`)
- Uses text/template for formatting
- Sections: Artifact Info, Size Breakdown, Top Files, Duplicates, Optimizations
- Human-readable with ASCII art and formatting

#### JSON Formatter (`internal/report/json.go`)
- Uses encoding/json with pretty printing
- Complete data structure serialization
- Supports both pretty and compact modes

#### Markdown Formatter (`internal/report/markdown.go`)
- GitHub-flavored markdown
- Tables for size breakdowns
- Collapsible sections for details
- Formatted for PR comments

#### HTML Formatter (`internal/report/html.go`)
- Embedded D3.js for treemap visualization
- Interactive size exploration
- Sortable tables
- Mobile-responsive design

## Contributing

### Contribution Workflow

1. **Fork** the repository
2. **Create feature branch**: `git checkout -b feature/my-feature`
3. **Make changes** with tests
4. **Format code**: `go fmt ./...`
5. **Run tests**: `go test ./...`
6. **Commit**: Follow conventional commits format
7. **Push**: `git push origin feature/my-feature`
8. **Pull request**: Submit PR with description

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build/tooling changes

**Examples**:
```
feat(ios): add XCArchive support

Implement analyzer for .xcarchive files to support Xcode archives
in addition to IPA files.

Closes #123

fix(analyzer): handle corrupted ZIP files

Add error handling for corrupted archive files to prevent panics.

docs(readme): update installation instructions

Add instructions for plugin installation via Bitrise CLI.
```

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Add comments for exported functions
- Keep functions focused and small
- Write descriptive variable names

### Pull Request Guidelines

1. **Title**: Clear, descriptive title
2. **Description**: Explain what and why
3. **Tests**: Include tests for new functionality
4. **Documentation**: Update README if needed
5. **Scope**: Keep PRs focused on single feature/fix

### Development Tips

1. **Incremental development**: Build features in small, testable chunks
2. **Test-driven**: Write tests before implementation when possible
3. **Use test artifacts**: Validate changes against real apps
4. **Profile performance**: Use `go test -bench` for performance-critical code
5. **Document decisions**: Add comments explaining non-obvious choices

### Getting Help

- **Issues**: Report bugs or request features via GitHub Issues
- **Discussions**: Ask questions in GitHub Discussions
- **Code review**: Request review from maintainers

## Advanced Development Topics

### Adding iOS Binary Analysis Features

To extend Mach-O parsing (e.g., add bitcode detection):

1. Extend `internal/analyzer/ios/macho/` package
2. Update binary metadata structures in `pkg/types/`
3. Add detection logic to parser
4. Update report formatters to display new data
5. Add tests with real binaries

### Adding Android Analysis Features

To extend Android analysis (e.g., parse AndroidManifest.xml):

1. Add parser to `internal/analyzer/android/`
2. Extract metadata during APK/AAB analysis
3. Add to report structure
4. Update formatters
5. Test with various APK structures

### Performance Optimization

For large artifacts:

1. **Profile**: Use `pprof` to identify bottlenecks
2. **Parallelize**: Use goroutines for independent operations
3. **Stream**: Process large files without loading fully into memory
4. **Cache**: Cache expensive computations (e.g., file hashes)
5. **Benchmark**: Use `go test -bench` to measure improvements

### Debugging Tips

```bash
# Enable verbose logging
DEBUG=1 ./bundle-inspector analyze app.ipa

# Profile memory usage
go tool pprof bundle-inspector profile.pb.gz

# Race condition detection
go test -race ./...

# Print analysis steps
./bundle-inspector analyze app.ipa -v
```

## Release Process

1. Update version in `main.go`
2. Update CHANGELOG.md
3. Tag release: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tags: `git push origin v1.0.0`
5. GitHub Actions builds and publishes binaries
6. Update plugin registry

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Mach-O File Format](https://developer.apple.com/library/archive/documentation/Performance/Conceptual/CodeFootprint/Articles/MachOOverview.html)
- [APK Format](https://en.wikipedia.org/wiki/Apk_(file_format))
- [AAB Format](https://developer.android.com/guide/app-bundle/app-bundle-format)
