# Bundle Inspector - Coding Standards & Conventions

This document outlines coding standards, architectural patterns, and best practices for the bundle-inspector project. These standards were established during a comprehensive cleanup in January 2026 and should be maintained going forward.

---

## Table of Contents

1. [Project Structure](#project-structure)
2. [Code Organization](#code-organization)
3. [Error Handling](#error-handling)
4. [Logging](#logging)
5. [Testing Standards](#testing-standards)
6. [Go Conventions](#go-conventions)
7. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)

---

## Project Structure

```
bundle-inspector/
├── cmd/
│   └── bundle-inspector/     # CLI entry point - keep minimal
├── pkg/
│   └── types/                # Public types (single source of truth)
├── internal/
│   ├── analyzer/             # Platform-specific analyzers (iOS, Android)
│   │   ├── ios/              # iOS analyzers (.ipa, .app)
│   │   └── android/          # Android analyzers (.apk, .aab)
│   ├── detector/             # Optimization detectors
│   ├── orchestrator/         # Business logic coordination
│   ├── logger/               # Structured logging
│   ├── util/                 # Shared utilities
│   └── testutil/             # Test helpers
└── test-artifacts/           # Sample files for testing
```

### Key Principles

- **Single Source of Truth**: Use `pkg/types` for all public type definitions
- **Separation of Concerns**: CLI layer (cmd) only handles I/O, business logic lives in orchestrator
- **No Import Cycles**: Place shared code in `util` package, not in analyzer/detector packages
- **Test Infrastructure**: Reusable test helpers go in `testutil`, not individual test files

---

## Code Organization

### Utility Functions

**DO**: Extract duplicated code into `internal/util/`

```go
// Good - shared utility in util package
func CalculateDiskUsage(fileSize int64) int64 {
    const BlockSize = 4096
    if fileSize == 0 {
        return 0
    }
    blocks := (fileSize + BlockSize - 1) / BlockSize
    return blocks * BlockSize
}
```

**DON'T**: Duplicate utility functions across files

```go
// Bad - duplicated in multiple detector files
func blockAlignedSize(size int64) int64 { /* same logic */ }
func calculateDiskUsage(size int64) int64 { /* same logic */ }
func calculateDiskUsageForImage(size int64) int64 { /* same logic */ }
```

### Business Logic

**DO**: Keep business logic in `internal/orchestrator/` or specialized packages

**DON'T**: Put business logic in `cmd/bundle-inspector/main.go`

The CLI should only:
- Parse flags
- Call orchestrator
- Format output
- Handle errors

Target: Keep `main.go` under 200 lines.

---

## Error Handling

### Detector Errors

**Always** wrap errors from detectors with context using `WrapError`:

```go
// Good - provides context
func (d *ImageOptimizationDetector) Detect(rootPath string) ([]types.Optimization, error) {
    if err := checkSipsAvailable(); err != nil {
        return nil, err // Already wrapped in checkSipsAvailable
    }

    err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
        // ... detector logic
        return nil
    })

    if err != nil {
        return nil, WrapError("image-optimization", "detecting optimizations", err)
    }

    return optimizations, nil
}

// Helper functions also wrap errors
func measureActualHEICConversion(imagePath string) (int64, error) {
    originalInfo, err := os.Stat(imagePath)
    if err != nil {
        return 0, WrapError("image-optimization", "measuring HEIC conversion",
            fmt.Errorf("failed to stat original: %w", err))
    }
    // ... rest of function
}
```

**DON'T**: Return raw errors without context

```go
// Bad - no context about which detector failed
func (d *DetectorName) Detect(rootPath string) ([]types.Optimization, error) {
    if err := someOperation(); err != nil {
        return nil, err // No context!
    }
}
```

### Error Context Components

When wrapping errors, provide:
1. **Detector/Component Name**: Identifies where the error originated
2. **Operation**: What was being attempted
3. **Original Error**: Preserved via error wrapping

Example error output:
```
image-optimization detector: measuring HEIC conversion: failed to stat original: no such file or directory
^----------------------^      ^-----------------------^  ^----------------------------------------------^
    detector name              operation                          original error
```

---

## Logging

### Structured Logger

**Always** use the structured logger from `internal/logger`:

```go
// Good - structured logging
type Orchestrator struct {
    Logger logger.Logger
    // ... other fields
}

func (o *Orchestrator) RunAnalysis(...) error {
    if err := o.runDetectors(...); err != nil {
        o.Logger.Warn("detector execution had issues: %v", err)
    }
}
```

**DON'T**: Use ad-hoc logging

```go
// Bad - ad-hoc logging
fmt.Fprintf(os.Stderr, "Warning: something failed: %v\n", err)
log.Printf("Error: %v", err)
```

### Log Levels

- **Debug**: Detailed diagnostic information (development only)
- **Info**: General informational messages (default level)
- **Warn**: Warning messages that don't stop execution
- **Error**: Errors that stop the current operation

### Logger Injection

Inject logger as a dependency for testability:

```go
// Production
orchestrator := orchestrator.New()
orchestrator.Logger = logger.NewDefaultLogger(os.Stderr, logger.LevelInfo)

// Testing
orchestrator := orchestrator.New()
orchestrator.Logger = logger.NewSilentLogger()
```

---

## Testing Standards

### Coverage Targets

- **New Detectors**: 80%+ coverage
- **Analyzers**: 60%+ coverage
- **Utilities**: 90%+ coverage
- **Overall Project**: 50%+ coverage

### Test Organization

**DO**: Use table-driven tests

```go
func TestValidateArtifact(t *testing.T) {
    tests := []struct {
        name      string
        setupPath func() string
        wantErr   bool
    }{
        {
            name: "valid APK file",
            setupPath: func() string {
                return testutil.CreateTestFile(t, tmpDir, "test.apk", 100)
            },
            wantErr: false,
        },
        {
            name: "invalid extension",
            setupPath: func() string {
                return testutil.CreateTestFile(t, tmpDir, "test.txt", 100)
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            path := tt.setupPath()
            err := ValidateArtifact(path)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateArtifact() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Test Helpers

**DO**: Create reusable test helpers in `internal/testutil/`

```go
// Good - reusable helpers
func CreateTempDir(t *testing.T) string
func CreateTestFile(t *testing.T, dir, filename string, size int64) string
func CreatePNGFile(t *testing.T, dir, filename string) string
```

**DON'T**: Duplicate test setup across test files

### Test File Naming

- Test files: `*_test.go` in the same package as the code
- Test helpers: `internal/testutil/testutil.go`
- Unique filenames in tests to avoid collisions (e.g., `test_dir.apk` not `test.apk` twice)

---

## Go Conventions

### Naming

- **Packages**: Short, lowercase, no underscores (`detector`, not `detector_utils`)
- **Files**: Lowercase with underscores (`image_optimization.go`)
- **Types**: PascalCase (`ImageOptimization`)
- **Functions/Methods**: camelCase (`detectPatterns`)
- **Constants**: PascalCase (`BlockSize`) or ALL_CAPS for acronyms (`KB`)

### Code Style

- **Line Length**: Aim for 100 characters, max 120
- **Function Length**: Keep under 50 lines; extract if longer
- **Complexity**: Avoid deeply nested logic (max 4 levels)

### Imports

Group imports in this order:
1. Standard library
2. External dependencies
3. Internal packages

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/external/package"

    "github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/analyzer"
    "github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)
```

### Error Handling

- Use `%w` for error wrapping: `fmt.Errorf("operation failed: %w", err)`
- Check errors immediately: don't defer error checking
- Return early on errors to reduce nesting

---

## Anti-Patterns to Avoid

### 1. Code Duplication

❌ **DON'T** duplicate utility functions across files
- Extract to `internal/util/`
- Create once, use everywhere

### 2. Type Duplication

❌ **DON'T** create multiple definitions of the same type
- Use `pkg/types` as single source of truth
- Don't create local types and then convert between them

### 3. Business Logic in CLI

❌ **DON'T** put business logic in `main.go`
- CLI should be thin: parse flags → call orchestrator → format output
- Keep `main.go` under 200 lines

### 4. Uncontextualized Errors

❌ **DON'T** return raw errors without context
- Always use `WrapError` for detector errors
- Include operation name in error messages

### 5. Ad-hoc Logging

❌ **DON'T** use `fmt.Fprintf(os.Stderr, ...)` or `log.Printf`
- Use the structured logger from `internal/logger`
- Inject logger for testability

### 6. Untested Code

❌ **DON'T** skip tests for new features
- New detectors require 80%+ coverage
- Use `testutil` helpers for consistent test setup

### 7. Validation Duplication

❌ **DON'T** duplicate validation logic
- Use `util.ValidateFileArtifact()` and `util.ValidateDirectoryArtifact()`
- Applies to all analyzers (.ipa, .app, .apk, .aab)

### 8. Unused Code

❌ **DON'T** leave unused code in the repository
- Remove immediately if not used
- No "commented out" code for "later"

---

## Architecture Patterns

### Orchestrator Pattern

The orchestrator coordinates the analysis workflow:

```
main.go (CLI)
    ↓
Orchestrator (business logic)
    ↓
Analyzer (platform-specific analysis)
    ↓
Detectors (optimization detection)
    ↓
Report (results)
```

### Analyzer Interface

All analyzers implement:
```go
type Analyzer interface {
    ValidateArtifact(path string) error
    Analyze(ctx context.Context, path string) (*types.Report, error)
}
```

### Detector Interface

All detectors implement:
```go
type Detector interface {
    Name() string
    Detect(rootPath string) ([]types.Optimization, error)
}
```

---

## Performance Considerations

### File Operations

- Use `filepath.Walk` for directory traversal
- Calculate disk-aligned sizes for accurate space reporting (4KB blocks on iOS)
- Parallelize expensive operations (e.g., hash computation in duplicate detector)

### Memory

- Stream large files when possible
- Clean up temporary files with `defer os.RemoveAll()`
- Don't load entire ZIP files into memory

---

## Documentation

### Code Comments

- Document **why**, not **what** (the code shows what)
- Document exported functions with godoc format
- Include examples for complex functions

```go
// CalculateDiskUsage returns the actual disk space used by a file.
// Files are stored in 4KB blocks on APFS (iOS), so a 95-byte file
// uses 4096 bytes on disk.
func CalculateDiskUsage(fileSize int64) int64 {
    // ...
}
```

### README and Documentation

- Keep README up-to-date with usage examples
- Document new features in commit messages
- Use Co-Authored-By for AI pair programming commits

---

## Git Conventions

### Commit Messages

Format:
```
<type>: <subject>

<body>

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

Types:
- `feat:` New feature
- `fix:` Bug fix
- `refactor:` Code refactoring
- `test:` Adding tests
- `docs:` Documentation changes
- `chore:` Maintenance tasks

### Branch Strategy

- `main`: Production-ready code
- Feature branches: Short-lived, descriptive names
- Clean commit history: Squash if needed

---

## Code Review Checklist

Before committing, verify:

- [ ] All tests pass (`go test ./...`)
- [ ] Build succeeds (`go build ./cmd/bundle-inspector`)
- [ ] Coverage targets met for new code
- [ ] Errors wrapped with context
- [ ] Logger used instead of fmt.Fprintf
- [ ] No code duplication
- [ ] Utilities in correct package
- [ ] Types use `pkg/types` (no local copies)
- [ ] Documentation added/updated
- [ ] No unused code

---

## Maintenance

### Regular Tasks

1. **Monitor Coverage**: Run `go test -cover ./...` regularly
2. **Check for Duplication**: Use tools like `gocyclo` and `dupl`
3. **Update Dependencies**: Keep external packages current
4. **Review Logs**: Ensure logging is useful and not excessive

### When Adding New Features

1. **Start with Tests**: Write tests first (TDD)
2. **Check for Existing Utilities**: Don't recreate what exists
3. **Use Established Patterns**: Follow existing detector/analyzer patterns
4. **Document**: Add godoc comments
5. **Update This File**: If introducing new patterns

---

## Questions?

If unsure about standards:
1. Look at existing code (especially detectors and analyzers)
2. Check this document
3. Refer to cleanup commits (January 2026) for examples
4. Maintain consistency with existing patterns

**Last Updated**: January 2026 (post-cleanup)
**Standards Version**: 1.0
