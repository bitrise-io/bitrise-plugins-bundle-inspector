package util

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSize parses a size string like "1MB", "500KB", "1.5GB" into bytes
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Extract number and unit
	var numStr string
	var unit string

	for i, c := range sizeStr {
		if c >= '0' && c <= '9' || c == '.' {
			numStr += string(c)
		} else {
			unit = strings.TrimSpace(sizeStr[i:])
			break
		}
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", sizeStr)
	}

	// Convert to bytes
	multiplier := int64(1)
	switch unit {
	case "B", "":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return int64(num * float64(multiplier)), nil
}
