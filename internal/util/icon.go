package util

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder

	"github.com/andrianbdn/iospng"
)

// IconSearchHints provides context from Info.plist to guide icon extraction.
type IconSearchHints struct {
	// PlistIconNames are base names from CFBundleIcons/CFBundleIconFiles
	// e.g., ["AppIcon60x60", "Icon-Production", "Telegram"]
	PlistIconNames []string
}

// ExtractIconFromZip extracts the app icon from a ZIP archive (IPA, APK, AAB)
// Returns base64-encoded data URI (e.g., "data:image/png;base64,...")
func ExtractIconFromZip(zipPath string, artifactType string) (string, error) {
	return ExtractIconFromZipWithHints(zipPath, artifactType, nil)
}

// ExtractIconFromZipWithHints extracts the app icon using Info.plist hints for better matching.
func ExtractIconFromZipWithHints(zipPath string, artifactType string, hints *IconSearchHints) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	var iconData []byte
	switch artifactType {
	case "ipa", "app":
		iconData, err = extractIOSIconWithHints(r, hints)
	case "apk", "aab":
		iconData, err = extractAndroidIcon(r)
	default:
		return "", fmt.Errorf("unsupported artifact type: %s", artifactType)
	}

	if err != nil {
		return "", err
	}

	if len(iconData) == 0 {
		return "", fmt.Errorf("no icon found")
	}

	// Convert to base64 data URI
	encoded := base64.StdEncoding.EncodeToString(iconData)
	return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
}

// ExtractIconFromDirectory extracts the app icon from a directory (.app bundle)
// Returns base64-encoded data URI (e.g., "data:image/png;base64,...")
func ExtractIconFromDirectory(dirPath string) (string, error) {
	return ExtractIconFromDirectoryWithHints(dirPath, nil)
}

// ExtractIconFromDirectoryWithHints extracts the app icon using Info.plist hints for better matching.
func ExtractIconFromDirectoryWithHints(dirPath string, hints *IconSearchHints) (string, error) {
	type candidate struct {
		path     string
		priority int
	}

	var candidates []candidate

	// Strategy 1: Info.plist-guided search (highest priority)
	if hints != nil {
		for _, baseName := range hints.PlistIconNames {
			matches, _ := filepath.Glob(filepath.Join(dirPath, baseName+"*.png"))
			for _, m := range matches {
				candidates = append(candidates, candidate{
					path:     m,
					priority: getIconPriority(filepath.Base(m)) + 50,
				})
			}
		}
	}

	// Strategy 2: AppIcon prefix (current behavior)
	matches, _ := filepath.Glob(filepath.Join(dirPath, "AppIcon*.png"))
	for _, m := range matches {
		candidates = append(candidates, candidate{
			path:     m,
			priority: getIconPriority(filepath.Base(m)),
		})
	}

	// Strategy 3: Broad heuristic — any PNG containing "icon" (case-insensitive)
	if len(candidates) == 0 {
		allPNGs, _ := filepath.Glob(filepath.Join(dirPath, "*.png"))
		for _, m := range allPNGs {
			name := filepath.Base(m)
			if strings.Contains(strings.ToLower(name), "icon") {
				candidates = append(candidates, candidate{
					path:     m,
					priority: 1,
				})
			}
		}
	}

	// Sort by priority (higher first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].priority > candidates[j].priority
	})

	// Try to extract the best candidate
	for _, c := range candidates {
		data, err := os.ReadFile(c.path)
		if err != nil {
			continue
		}

		if isCgBIPNG(data) {
			data, err = convertCgBIToStandardPNG(data)
			if err != nil {
				continue
			}
		}

		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			continue
		}

		encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
		return fmt.Sprintf("data:image/png;base64,%s", encoded), nil
	}

	return "", fmt.Errorf("no icon found in directory")
}

// extractIOSIcon extracts icon from iOS IPA (backward-compatible wrapper).
func extractIOSIcon(r *zip.ReadCloser) ([]byte, error) {
	return extractIOSIconWithHints(r, nil)
}

// extractIOSIconWithHints extracts icon from iOS IPA using a multi-strategy approach:
//  1. Info.plist-guided: match PNGs whose name starts with a declared icon base name (+50 priority bonus)
//  2. AppIcon prefix: existing behavior for apps using standard naming
//  3. Broad heuristic: any PNG in .app root containing "icon" (lowest priority)
func extractIOSIconWithHints(r *zip.ReadCloser, hints *IconSearchHints) ([]byte, error) {
	type candidate struct {
		file     *zip.File
		priority int
		size     int
	}

	var candidates []candidate

	for _, file := range r.File {
		name := filepath.Base(file.Name)
		dir := filepath.Dir(file.Name)

		// Must be in a .app directory and be a PNG
		if !strings.Contains(dir, ".app") || !strings.HasSuffix(strings.ToLower(name), ".png") {
			continue
		}

		// Skip files deep inside Frameworks/ or PlugIns/ subdirectories
		if strings.Contains(dir, "/Frameworks/") || strings.Contains(dir, "/PlugIns/") {
			continue
		}

		matched := false
		baseName := strings.TrimSuffix(name, filepath.Ext(name))

		// Strategy 1: Info.plist-guided search
		if hints != nil {
			for _, plistName := range hints.PlistIconNames {
				if strings.HasPrefix(baseName, plistName) {
					candidates = append(candidates, candidate{
						file:     file,
						priority: getIconPriority(name) + 50,
						size:     int(file.UncompressedSize64),
					})
					matched = true
					break
				}
			}
		}

		if matched {
			continue
		}

		// Strategy 2: AppIcon prefix (current behavior)
		if strings.HasPrefix(name, "AppIcon") {
			candidates = append(candidates, candidate{
				file:     file,
				priority: getIconPriority(name),
				size:     int(file.UncompressedSize64),
			})
			continue
		}

		// Strategy 3: Broad heuristic — any PNG containing "icon" in name
		if strings.Contains(strings.ToLower(name), "icon") {
			candidates = append(candidates, candidate{
				file:     file,
				priority: 1,
				size:     int(file.UncompressedSize64),
			})
		}
	}

	// Sort by priority (higher first), then by size (larger first)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority > candidates[j].priority
		}
		return candidates[i].size > candidates[j].size
	})

	// Try to extract the best candidate
	for _, c := range candidates {
		data, err := extractAndConvertIcon(c.file)
		if err == nil && len(data) > 0 {
			return data, nil
		}
	}

	return nil, fmt.Errorf("no valid icon found")
}

// extractAndroidIcon extracts icon from Android APK/AAB
func extractAndroidIcon(r *zip.ReadCloser) ([]byte, error) {
	// Look for ic_launcher.png in various densities
	// Priority: xxxhdpi > xxhdpi > xhdpi > hdpi > mdpi

	densities := []string{
		"mipmap-xxxhdpi",
		"mipmap-xxhdpi",
		"mipmap-xhdpi",
		"mipmap-hdpi",
		"drawable-xxxhdpi",
		"drawable-xxhdpi",
		"drawable-xhdpi",
		"drawable-hdpi",
		"mipmap-mdpi",
		"drawable-mdpi",
		"mipmap",
		"drawable",
	}

	iconNames := []string{
		"ic_launcher.png",
		"ic_launcher_round.png",
		"icon.png",
	}

	// Try each density in order
	for _, density := range densities {
		for _, iconName := range iconNames {
			path := fmt.Sprintf("res/%s/%s", density, iconName)

			for _, file := range r.File {
				if file.Name == path {
					data, err := extractAndConvertIcon(file)
					if err == nil && len(data) > 0 {
						return data, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no valid icon found")
}

// extractAndConvertIcon reads an icon file and converts it to PNG if needed
func extractAndConvertIcon(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open icon file: %w", err)
	}
	defer rc.Close()

	// Read the file data
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read icon data: %w", err)
	}

	// Check if it's a CgBI PNG and convert if needed
	if isCgBIPNG(data) {
		data, err = convertCgBIToStandardPNG(data)
		if err != nil {
			return nil, fmt.Errorf("failed to convert CgBI PNG: %w", err)
		}
	}

	// Decode the image (now standard PNG if it was CgBI)
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// If already PNG and size is reasonable, return as-is
	if format == "png" && len(data) < 200*1024 { // Less than 200KB
		return data, nil
	}

	// Convert to PNG and/or resize if needed
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is too large (> 512px), we might want to resize
	// For now, just re-encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	// Verify size is reasonable
	if width == 0 || height == 0 {
		return nil, fmt.Errorf("invalid image dimensions")
	}

	return buf.Bytes(), nil
}

// getIconPriority returns priority for iOS icon names (higher = better)
func getIconPriority(name string) int {
	name = strings.ToLower(name)

	// Prefer 3x over 2x over 1x
	if strings.Contains(name, "@3x") {
		return 30
	}
	if strings.Contains(name, "@2x") {
		return 20
	}
	if strings.Contains(name, "@1x") {
		return 10
	}

	// Prefer 60x60 (iPhone) icons
	if strings.Contains(name, "60x60") {
		return 25
	}

	// General AppIcon
	if strings.HasPrefix(name, "appicon") {
		return 5
	}

	return 1
}

// isCgBIPNG checks if data is a CgBI-optimized PNG file.
// CgBI is Apple's proprietary PNG format used in iOS apps with optimizations like:
// - BGR color order (instead of RGB)
// - Premultiplied alpha channel
// - Apple-specific compression
//
// CgBI PNGs have the standard PNG signature followed by a CgBI chunk before IHDR.
func isCgBIPNG(data []byte) bool {
	// Need at least PNG signature (8 bytes) + chunk length (4) + "CgBI" (4)
	if len(data) < 16 {
		return false
	}

	// Check PNG signature: 89 50 4E 47 0D 0A 1A 0A
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.Equal(data[0:8], pngSignature) {
		return false
	}

	// Check for CgBI chunk in the first 20 bytes after signature
	// CgBI chunk typically appears at offset 12 (after 8-byte signature + 4-byte chunk length)
	return bytes.Contains(data[8:20], []byte("CgBI"))
}

// convertCgBIToStandardPNG converts Apple's CgBI PNG format to standard PNG.
// Uses the iospng library to handle the conversion, which:
// 1. Reverts Apple's PNG optimizations (BGR → RGB, un-premultiply alpha)
// 2. Produces a standard PNG that can be decoded by image.Decode
func convertCgBIToStandardPNG(data []byte) ([]byte, error) {
	// Use iospng to revert Apple's PNG optimization
	reader := bytes.NewReader(data)
	var buf bytes.Buffer

	err := iospng.PngRevertOptimization(reader, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to revert CgBI PNG optimization: %w", err)
	}

	return buf.Bytes(), nil
}
