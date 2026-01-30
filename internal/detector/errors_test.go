package detector

import (
	"errors"
	"strings"
	"testing"
)

func TestDetectorError_Error(t *testing.T) {
	baseErr := errors.New("base error")
	detectorErr := &DetectorError{
		DetectorName: "test-detector",
		Operation:    "scanning",
		Err:          baseErr,
	}

	expected := "test-detector detector: scanning: base error"
	if detectorErr.Error() != expected {
		t.Errorf("Error() = %q, want %q", detectorErr.Error(), expected)
	}
}

func TestDetectorError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	detectorErr := &DetectorError{
		DetectorName: "test-detector",
		Operation:    "scanning",
		Err:          baseErr,
	}

	if detectorErr.Unwrap() != baseErr {
		t.Error("Unwrap() did not return the base error")
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name         string
		detectorName string
		operation    string
		err          error
		wantNil      bool
		wantContains []string
	}{
		{
			name:         "nil error returns nil",
			detectorName: "test",
			operation:    "op",
			err:          nil,
			wantNil:      true,
		},
		{
			name:         "wraps error with context",
			detectorName: "duplicate",
			operation:    "detecting",
			err:          errors.New("file not found"),
			wantNil:      false,
			wantContains: []string{"duplicate", "detecting", "file not found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapError(tt.detectorName, tt.operation, tt.err)

			if tt.wantNil {
				if got != nil {
					t.Errorf("WrapError() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Error("WrapError() = nil, want error")
				return
			}

			errStr := got.Error()
			for _, want := range tt.wantContains {
				if !strings.Contains(errStr, want) {
					t.Errorf("Error string %q does not contain %q", errStr, want)
				}
			}

			// Test unwrapping
			if !errors.Is(got, tt.err) {
				t.Error("errors.Is() = false, wrapped error should match original")
			}
		})
	}
}
