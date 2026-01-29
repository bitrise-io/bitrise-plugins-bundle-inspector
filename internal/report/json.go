// Package report provides output formatters for analysis reports.
package report

import (
	"encoding/json"
	"io"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// JSONFormatter formats reports as JSON.
type JSONFormatter struct {
	indent bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(indent bool) *JSONFormatter {
	return &JSONFormatter{
		indent: indent,
	}
}

// Format writes the report in JSON format to the writer.
func (f *JSONFormatter) Format(w io.Writer, report *types.Report) error {
	encoder := json.NewEncoder(w)

	if f.indent {
		encoder.SetIndent("", "  ")
	}

	return encoder.Encode(report)
}
