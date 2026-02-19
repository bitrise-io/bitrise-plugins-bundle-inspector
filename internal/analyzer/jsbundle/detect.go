// Package jsbundle provides detection and analysis of JavaScript bundles
// found in React Native mobile applications.
package jsbundle

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
)

// BundleFormat represents the format of a JavaScript bundle.
type BundleFormat string

const (
	FormatMetro     BundleFormat = "metro"
	FormatHermes    BundleFormat = "hermes"
	FormatRAMBundle BundleFormat = "ram_bundle"
	FormatUnknown   BundleFormat = "unknown"
)

// Magic bytes for bundle format identification.
const (
	HermesMagic    uint32 = 0xC061D00D
	RAMBundleMagic uint32 = 0xFB0BD1E5
)

// IsJSBundleFilename checks if a filename is a known JavaScript bundle file.
// This is the zero-overhead check used to skip non-React Native apps.
func IsJSBundleFilename(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".jsbundle") || lower == "index.android.bundle"
}

// DetectFormat reads the first bytes of a reader to determine the bundle format.
// It checks for Hermes bytecode and RAM bundle magic bytes first, then scans
// for Metro module patterns in plain text bundles.
func DetectFormat(r io.ReaderAt) (BundleFormat, error) {
	// Read first 4 bytes for magic number detection
	magicBuf := make([]byte, 4)
	n, err := r.ReadAt(magicBuf, 0)
	if err != nil && n < 4 {
		if n == 0 {
			return FormatUnknown, nil
		}
		return FormatUnknown, nil
	}

	magic := binary.LittleEndian.Uint32(magicBuf)

	switch magic {
	case HermesMagic:
		return FormatHermes, nil
	case RAMBundleMagic:
		return FormatRAMBundle, nil
	}

	// Not a binary format — scan for Metro module patterns in text
	textBuf := make([]byte, 512)
	n, _ = r.ReadAt(textBuf, 0)
	if n == 0 {
		return FormatUnknown, nil
	}

	text := string(textBuf[:n])
	if strings.Contains(text, "__d(function") || strings.Contains(text, "__d(") {
		return FormatMetro, nil
	}
	if bytes.HasPrefix(textBuf[:n], []byte("var ")) {
		return FormatMetro, nil
	}

	return FormatUnknown, nil
}
