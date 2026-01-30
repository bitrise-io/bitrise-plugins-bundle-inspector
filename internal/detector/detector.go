package detector

import (
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// Detector is the interface for all optimization detectors
type Detector interface {
	// Detect runs the detector and returns optimization recommendations
	Detect(rootPath string) ([]types.Optimization, error)

	// Name returns the detector's name for logging
	Name() string
}
