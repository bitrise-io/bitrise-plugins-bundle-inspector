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
)

// ExtractIconFromZip extracts the app icon from a ZIP archive (IPA, APK, AAB)
// Returns base64-encoded data URI (e.g., "data:image/png;base64,...")
func ExtractIconFromZip(zipPath string, artifactType string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer r.Close()

	var iconData []byte
	switch artifactType {
	case "ipa", "app":
		iconData, err = extractIOSIcon(r)
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
	// Look for AppIcon files
	iconNames := []string{
		"AppIcon60x60@3x.png",
		"AppIcon60x60@2x.png",
		"AppIcon76x76@2x~ipad.png",
		"AppIcon60x60@1x.png",
	}

	for _, iconName := range iconNames {
		iconPath := filepath.Join(dirPath, iconName)
		if data, err := os.ReadFile(iconPath); err == nil {
			// Convert to PNG if needed and encode
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
	}

	return "", fmt.Errorf("no icon found in directory")
}

// extractIOSIcon extracts icon from iOS IPA
func extractIOSIcon(r *zip.ReadCloser) ([]byte, error) {
	// Look for AppIcon files in Payload/*.app/
	// Priority: AppIcon60x60@3x.png > AppIcon60x60@2x.png > AppIcon*.png

	var candidates []struct {
		file     *zip.File
		priority int
		size     int
	}

	for _, file := range r.File {
		name := filepath.Base(file.Name)
		dir := filepath.Dir(file.Name)

		// Check if it's in the .app directory
		if !strings.Contains(dir, ".app") {
			continue
		}

		// Look for AppIcon files
		if strings.HasPrefix(name, "AppIcon") && strings.HasSuffix(name, ".png") {
			priority := getIconPriority(name)
			candidates = append(candidates, struct {
				file     *zip.File
				priority int
				size     int
			}{
				file:     file,
				priority: priority,
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
	for _, candidate := range candidates {
		data, err := extractAndConvertIcon(candidate.file)
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

	// Decode the image
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
