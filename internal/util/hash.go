// Package util provides utility functions for the bundle inspector.
package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// ComputeSHA256 computes the SHA-256 hash of a file using chunked reading.
func ComputeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	// Use 64KB chunks for memory efficiency
	buf := make([]byte, 64*1024)

	if _, err := io.CopyBuffer(h, f, buf); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
