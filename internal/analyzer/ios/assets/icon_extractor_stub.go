//go:build !darwin

package assets

import "fmt"

// ExtractIconFromCar is a stub for non-macOS systems.
// Assets.car icon extraction requires macOS AppKit framework.
func ExtractIconFromCar(carPath string, iconNames []string, catalogAssets []AssetInfo) ([]byte, error) {
	return nil, fmt.Errorf("Assets.car icon extraction requires macOS")
}
