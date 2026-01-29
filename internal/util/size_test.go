package util

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s; want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		part     int64
		total    int64
		expected string
	}{
		{0, 100, "0.0%"},
		{50, 100, "50.0%"},
		{33, 100, "33.0%"},
		{100, 100, "100.0%"},
		{0, 0, "0.0%"},
		{25, 200, "12.5%"},
	}

	for _, tt := range tests {
		result := FormatPercentage(tt.part, tt.total)
		if result != tt.expected {
			t.Errorf("FormatPercentage(%d, %d) = %s; want %s",
				tt.part, tt.total, result, tt.expected)
		}
	}
}
