//go:build darwin

package assets

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// swiftIconScript is a small Swift program that extracts an icon from an Assets.car
// file by loading it as an NSBundle resource and writing PNG data to stdout.
// Uses only public Apple APIs (NSBundle, NSImage, NSBitmapImageRep).
//
// Accepts: <bundle-path> <icon-name-1> [icon-name-2] ... [icon-name-N]
// Tries each icon name in order and outputs the first one found.
const swiftIconScript = `import AppKit

guard CommandLine.arguments.count >= 3 else {
    FileHandle.standardError.write("usage: <bundle-path> <icon-name> [...]\n".data(using: .utf8)!)
    exit(1)
}
let bundlePath = CommandLine.arguments[1]
let iconNames = Array(CommandLine.arguments[2...])

guard let bundle = Bundle(path: bundlePath) else {
    FileHandle.standardError.write("failed to load bundle at \(bundlePath)\n".data(using: .utf8)!)
    exit(1)
}

for iconName in iconNames {
    guard let image = bundle.image(forResource: iconName) else { continue }

    // Get the best representation (largest)
    var bestRep: NSBitmapImageRep?
    for rep in image.representations {
        if let bitmapRep = rep as? NSBitmapImageRep {
            if bestRep == nil || (bitmapRep.pixelsWide * bitmapRep.pixelsHigh) > (bestRep!.pixelsWide * bestRep!.pixelsHigh) {
                bestRep = bitmapRep
            }
        }
    }

    // Fall back to CGImage conversion if no bitmap rep found
    if bestRep == nil {
        guard let cgImage = image.cgImage(forProposedRect: nil, context: nil, hints: nil) else { continue }
        bestRep = NSBitmapImageRep(cgImage: cgImage)
    }

    guard let rep = bestRep,
          let pngData = rep.representation(using: .png, properties: [:])
    else { continue }

    FileHandle.standardOutput.write(pngData)
    exit(0)
}

FileHandle.standardError.write("no icon found for names: \(iconNames.joined(separator: ", "))\n".data(using: .utf8)!)
exit(1)
`

// swiftTimeout is the maximum time allowed for the Swift helper to compile and run.
const swiftTimeout = 60 * time.Second

// pngSignature is the magic bytes at the start of every valid PNG file.
var pngSignature = [8]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

// ExtractIconFromCar extracts an icon PNG from an Assets.car file.
// It tries each candidate name in order: plist icon names, icon-type assets from
// the parsed catalog, and common fallback names.
// Returns raw PNG bytes, or error if extraction fails.
func ExtractIconFromCar(ctx context.Context, carPath string, iconNames []string, catalogAssets []AssetInfo) ([]byte, error) {
	// Build deduplicated candidate list
	seen := make(map[string]struct{})
	var candidates []string

	// Priority 1: Names from Info.plist
	for _, name := range iconNames {
		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			candidates = append(candidates, name)
		}
	}

	// Priority 2: Icon-type assets from parsed catalog
	for _, asset := range catalogAssets {
		if asset.Type == "icon" && asset.Name != "" {
			if _, exists := seen[asset.Name]; !exists {
				seen[asset.Name] = struct{}{}
				candidates = append(candidates, asset.Name)
			}
		}
	}

	// Priority 3: Common fallback names
	for _, name := range []string{"AppIcon", "app_icon", "Icon"} {
		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no icon candidate names to try")
	}

	return extractIconWithSwift(ctx, carPath, candidates)
}

// extractIconWithSwift extracts a named icon from an Assets.car file by creating
// a temporary bundle structure and running a Swift script that tries all candidate
// names in a single invocation.
func extractIconWithSwift(ctx context.Context, carPath string, iconNames []string) ([]byte, error) {
	// Create a temporary bundle directory: <tmp>/Contents/Resources/
	bundleDir, err := os.MkdirTemp("", "icon-bundle-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(bundleDir)

	resourcesDir := filepath.Join(bundleDir, "Contents", "Resources")
	if err := os.MkdirAll(resourcesDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create bundle structure: %w", err)
	}

	// Symlink Assets.car into the bundle's Resources directory
	absCarPath, err := filepath.Abs(carPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve car path: %w", err)
	}
	if err := os.Symlink(absCarPath, filepath.Join(resourcesDir, "Assets.car")); err != nil {
		return nil, fmt.Errorf("failed to symlink Assets.car: %w", err)
	}

	// Write the Swift script to a temp file
	scriptFile, err := os.CreateTemp("", "icon-extract-*.swift")
	if err != nil {
		return nil, fmt.Errorf("failed to create script file: %w", err)
	}
	defer os.Remove(scriptFile.Name())

	if _, err := scriptFile.WriteString(swiftIconScript); err != nil {
		scriptFile.Close()
		return nil, fmt.Errorf("failed to write script: %w", err)
	}
	scriptFile.Close()

	// Run: swift <script> <bundle-path> <icon-name-1> <icon-name-2> ...
	ctx, cancel := context.WithTimeout(ctx, swiftTimeout)
	defer cancel()

	args := append([]string{scriptFile.Name(), bundleDir}, iconNames...)
	cmd := exec.CommandContext(ctx, "swift", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("swift icon extraction failed: %w (stderr: %s)", err, stderrStr)
		}
		return nil, fmt.Errorf("swift icon extraction failed: %w", err)
	}

	data := stdout.Bytes()

	// Validate PNG output
	if len(data) < len(pngSignature) || !bytes.Equal(data[:len(pngSignature)], pngSignature[:]) {
		return nil, fmt.Errorf("swift output is not a valid PNG (%d bytes)", len(data))
	}

	return data, nil
}
