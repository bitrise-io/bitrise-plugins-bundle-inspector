// Package report provides output formatters for analysis reports.
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// TextFormatter formats reports as human-readable text.
type TextFormatter struct{}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

// Format writes the report in text format to the writer.
func (f *TextFormatter) Format(w io.Writer, report *types.Report) error {
	// Header
	fmt.Fprintf(w, "Bundle Inspector Analysis Report\n")
	fmt.Fprintf(w, "=================================\n\n")

	// Artifact Info
	fmt.Fprintf(w, "Artifact Information:\n")
	fmt.Fprintf(w, "  Type: %s\n", report.ArtifactInfo.Type)
	fmt.Fprintf(w, "  Path: %s\n", report.ArtifactInfo.Path)
	fmt.Fprintf(w, "  Compressed Size: %s\n", util.FormatBytes(report.ArtifactInfo.Size))
	if report.ArtifactInfo.UncompressedSize > 0 {
		fmt.Fprintf(w, "  Uncompressed Size: %s\n", util.FormatBytes(report.ArtifactInfo.UncompressedSize))
		ratio := float64(report.ArtifactInfo.Size) / float64(report.ArtifactInfo.UncompressedSize) * 100
		fmt.Fprintf(w, "  Compression Ratio: %.1f%%\n", ratio)
	}
	fmt.Fprintf(w, "\n")

	// Size Breakdown
	fmt.Fprintf(w, "Size Breakdown:\n")
	totalSize := report.ArtifactInfo.UncompressedSize
	if totalSize == 0 {
		totalSize = report.ArtifactInfo.Size
	}

	breakdown := []struct {
		name string
		size int64
	}{
		{"Executable", report.SizeBreakdown.Executable},
		{"Frameworks", report.SizeBreakdown.Frameworks},
		{"Libraries", report.SizeBreakdown.Libraries},
		{"Assets", report.SizeBreakdown.Assets},
		{"Resources", report.SizeBreakdown.Resources},
	}

	if report.SizeBreakdown.DEX > 0 {
		breakdown = append(breakdown, struct {
			name string
			size int64
		}{"DEX Files", report.SizeBreakdown.DEX})
	}

	if report.SizeBreakdown.JavaScript > 0 {
		breakdown = append(breakdown, struct {
			name string
			size int64
		}{"JavaScript", report.SizeBreakdown.JavaScript})
	}

	breakdown = append(breakdown, struct {
		name string
		size int64
	}{"Other", report.SizeBreakdown.Other})

	for _, item := range breakdown {
		if item.size > 0 {
			fmt.Fprintf(w, "  %s: %s (%s)\n",
				item.name,
				util.FormatBytes(item.size),
				util.FormatPercentage(item.size, totalSize))
		}
	}
	fmt.Fprintf(w, "\n")

	// Category Breakdown
	if len(report.SizeBreakdown.ByCategory) > 0 {
		fmt.Fprintf(w, "Detailed Breakdown by Category:\n")
		for category, size := range report.SizeBreakdown.ByCategory {
			fmt.Fprintf(w, "  %s: %s (%s)\n",
				category,
				util.FormatBytes(size),
				util.FormatPercentage(size, totalSize))
		}
		fmt.Fprintf(w, "\n")
	}

	// Largest Files
	if len(report.LargestFiles) > 0 {
		fmt.Fprintf(w, "Top %d Largest Files:\n", len(report.LargestFiles))
		for i, file := range report.LargestFiles {
			fmt.Fprintf(w, "  %2d. %s - %s (%s)\n",
				i+1,
				file.Path,
				util.FormatBytes(file.Size),
				util.FormatPercentage(file.Size, totalSize))
		}
		fmt.Fprintf(w, "\n")
	}

	// Duplicates
	if len(report.Duplicates) > 0 {
		fmt.Fprintf(w, "Duplicate Files:\n")
		totalWasted := int64(0)
		for _, dup := range report.Duplicates {
			totalWasted += dup.WastedSize
			fmt.Fprintf(w, "  %d copies of %s (%s each):\n",
				dup.Count,
				dup.Files[0],
				util.FormatBytes(dup.Size))
			for _, file := range dup.Files {
				fmt.Fprintf(w, "    - %s\n", file)
			}
			fmt.Fprintf(w, "    Wasted space: %s\n", util.FormatBytes(dup.WastedSize))
		}
		fmt.Fprintf(w, "  Total wasted space: %s\n\n", util.FormatBytes(totalWasted))
	}

	// Optimizations
	if len(report.Optimizations) > 0 {
		fmt.Fprintf(w, "Optimization Opportunities:\n")

		// Group by severity
		highSev := []types.Optimization{}
		mediumSev := []types.Optimization{}
		lowSev := []types.Optimization{}

		for _, opt := range report.Optimizations {
			switch opt.Severity {
			case "high":
				highSev = append(highSev, opt)
			case "medium":
				mediumSev = append(mediumSev, opt)
			case "low":
				lowSev = append(lowSev, opt)
			}
		}

		printOptimizations := func(title string, opts []types.Optimization) {
			if len(opts) == 0 {
				return
			}
			fmt.Fprintf(w, "\n  %s Priority:\n", title)
			for _, opt := range opts {
				fmt.Fprintf(w, "    • %s\n", opt.Title)
				fmt.Fprintf(w, "      %s\n", opt.Description)
				if opt.Impact > 0 {
					fmt.Fprintf(w, "      Potential savings: %s\n", util.FormatBytes(opt.Impact))
				}
				if opt.Action != "" {
					fmt.Fprintf(w, "      Action: %s\n", opt.Action)
				}
			}
		}

		printOptimizations("High", highSev)
		printOptimizations("Medium", mediumSev)
		printOptimizations("Low", lowSev)

		fmt.Fprintf(w, "\n")
	}

	// Summary
	if report.TotalSavings > 0 {
		fmt.Fprintf(w, "Total Potential Savings: %s (%s)\n",
			util.FormatBytes(report.TotalSavings),
			util.FormatPercentage(report.TotalSavings, totalSize))
	}

	return nil
}

// FormatToString formats the report as a string.
func FormatToString(report *types.Report) (string, error) {
	var sb strings.Builder
	formatter := NewTextFormatter()
	if err := formatter.Format(&sb, report); err != nil {
		return "", err
	}
	return sb.String(), nil
}
