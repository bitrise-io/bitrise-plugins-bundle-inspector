package orchestrator

import (
	"testing"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name      string
		impact    int64
		totalSize int64
		want      string
	}{
		{
			name:      "high severity - 15% impact",
			impact:    15 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "high",
		},
		{
			name:      "high severity - exactly 10% impact",
			impact:    10 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "high",
		},
		{
			name:      "medium severity - 7% impact",
			impact:    7 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "medium",
		},
		{
			name:      "medium severity - exactly 5% impact",
			impact:    5 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "medium",
		},
		{
			name:      "low severity - 2% impact",
			impact:    2 * 1024 * 1024,
			totalSize: 100 * 1024 * 1024,
			want:      "low",
		},
		{
			name:      "low severity - zero total",
			impact:    1 * 1024 * 1024,
			totalSize: 0,
			want:      "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSeverity(tt.impact, tt.totalSize)
			if got != tt.want {
				t.Errorf("getSeverity(%d, %d) = %s, want %s", tt.impact, tt.totalSize, got, tt.want)
			}
		})
	}
}

func TestCalculateTotalSavings(t *testing.T) {
	report := &types.Report{
		Optimizations: []types.Optimization{
			{Impact: 1024 * 1024},     // 1 MB
			{Impact: 2 * 1024 * 1024}, // 2 MB
			{Impact: 500 * 1024},      // 500 KB
			{Impact: 0},               // 0 bytes
		},
	}

	expected := int64(1024*1024 + 2*1024*1024 + 500*1024)
	got := calculateTotalSavings(report)

	if got != expected {
		t.Errorf("calculateTotalSavings() = %d, want %d", got, expected)
	}
}

func TestCalculateTotalSavings_EmptyOptimizations(t *testing.T) {
	report := &types.Report{
		Optimizations: []types.Optimization{},
	}

	expected := int64(0)
	got := calculateTotalSavings(report)

	if got != expected {
		t.Errorf("calculateTotalSavings() with empty optimizations = %d, want %d", got, expected)
	}
}

func TestNew(t *testing.T) {
	orch := New()

	if orch == nil {
		t.Fatal("New() returned nil")
	}

	if !orch.IncludeDuplicates {
		t.Error("Expected IncludeDuplicates to be true by default")
	}

	expectedThreshold := int64(1024 * 1024) // 1MB
	if orch.ThresholdBytes != expectedThreshold {
		t.Errorf("Expected ThresholdBytes to be %d, got %d", expectedThreshold, orch.ThresholdBytes)
	}
}
