package detector

import "fmt"

// DetectorError wraps errors from detectors with additional context
type DetectorError struct {
	DetectorName string
	Operation    string
	Err          error
}

// Error implements the error interface
func (e *DetectorError) Error() string {
	return fmt.Sprintf("%s detector: %s: %v", e.DetectorName, e.Operation, e.Err)
}

// Unwrap returns the wrapped error
func (e *DetectorError) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with detector context
func WrapError(detectorName, operation string, err error) error {
	if err == nil {
		return nil
	}
	return &DetectorError{
		DetectorName: detectorName,
		Operation:    operation,
		Err:          err,
	}
}
