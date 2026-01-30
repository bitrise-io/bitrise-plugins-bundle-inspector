package report

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/util"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

// MarkdownFormatter formats reports as GitHub/GitLab compatible markdown
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new markdown formatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// Format writes the report in markdown format to the writer
func (f *MarkdownFormatter) Format(w io.Writer, report *types.Report) error {
	if err := f.writeHeader(w, report); err != nil {
		return err
	}

	// Group optimizations by category
	categoryGroups := getCategoryGroups(report.Optimizations)

	// Write optimizations by category in order
	categories := []struct {
		key   string
		name  string
		emoji string
		open  bool
	}{
		{"strip-symbols", "Strip Binary Symbols", "🔧", true},
		{"duplicates", "Duplicate Files", "🔄", false},
		{"image-optimization", "Image Optimization", "🖼️", false},
		{"loose-images", "Loose Images", "📸", false},
		{"unnecessary-files", "Unnecessary Files", "🗑️", false},
	}

	for _, cat := range categories {
		if opts, exists := categoryGroups[cat.key]; exists && len(opts) > 0 {
			if err := f.writeOptimizations(w, opts, cat.name, cat.emoji, cat.open); err != nil {
				return err
			}
		}
	}

	// Write size breakdown
	if err := f.writeSizeBreakdown(w, report); err != nil {
		return err
	}

	// Write largest files
	if err := f.writeLargestFiles(w, report); err != nil {
		return err
	}

	// Write duplicates detail (only if present)
	if len(report.Duplicates) > 0 {
		if err := f.writeDuplicates(w, report); err != nil {
			return err
		}
	}

	// Write by extension (only if data exists)
	if err := f.writeByExtension(w, report); err != nil {
		return err
	}

	return nil
}

// writeHeader writes the always-visible header section
func (f *MarkdownFormatter) writeHeader(w io.Writer, report *types.Report) error {
	if _, err := fmt.Fprintf(w, "## Bitrise Report\n\n"); err != nil {
		return err
	}

	// Extract artifact name
	artifactName := report.ArtifactInfo.Path
	if idx := strings.LastIndex(artifactName, "/"); idx >= 0 {
		artifactName = artifactName[idx+1:]
	}

	// Get commit hash from metadata if available
	commitHash := "-"
	if report.Metadata != nil {
		if commit, ok := report.Metadata["commit_hash"].(string); ok && commit != "" {
			if len(commit) > 7 {
				commitHash = commit[:7]
			} else {
				commitHash = commit
			}
		}
	}

	// Calculate sizes
	uncompressedSize := calculateUncompressedSize(&report.SizeBreakdown)
	downloadSize := report.ArtifactInfo.Size
	installSize := uncompressedSize

	// Format potential savings
	savingsPercentage := 0.0
	if uncompressedSize > 0 {
		savingsPercentage = float64(report.TotalSavings) / float64(uncompressedSize) * 100
	}
	potentialSavings := fmt.Sprintf("%s (%.1f%%)", util.FormatBytes(report.TotalSavings), savingsPercentage)

	// Write horizontal summary table
	if _, err := fmt.Fprintf(w, "| Bundle | Commit | Install Size | Download Size | Potential Savings |\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "|--------|--------|--------------|---------------|-------------------|\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "| %s | `%s` | %s | %s | %s |\n\n",
		artifactName, commitHash, util.FormatBytes(installSize), util.FormatBytes(downloadSize), potentialSavings); err != nil {
		return err
	}

	// Additional context
	totalFiles := countFiles(report.FileTree)
	categoryGroups := getCategoryGroups(report.Optimizations)
	if _, err := fmt.Fprintf(w, "**Analysis:** %s files analyzed, %d optimizations found across %d categories\n\n",
		util.FormatNumber(int64(totalFiles)), len(report.Optimizations), len(categoryGroups)); err != nil {
		return err
	}

	return nil
}

// writeSizeBreakdown writes the size breakdown by category section
func (f *MarkdownFormatter) writeSizeBreakdown(w io.Writer, report *types.Report) error {
	breakdown := map[string]int64{
		"Frameworks":  report.SizeBreakdown.Frameworks,
		"Resources":   report.SizeBreakdown.Resources,
		"Executable":  report.SizeBreakdown.Executable,
		"Assets":      report.SizeBreakdown.Assets,
		"Libraries":   report.SizeBreakdown.Libraries,
		"DEX":         report.SizeBreakdown.DEX,
		"Other":       report.SizeBreakdown.Other,
	}

	// Filter out zero values
	filtered := make(map[string]int64)
	for k, v := range breakdown {
		if v > 0 {
			filtered[k] = v
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	sorted := sortBySize(filtered)
	total := calculateUncompressedSize(&report.SizeBreakdown)

	if _, err := fmt.Fprintf(w, "<details>\n<summary><strong>📊 Size Breakdown by Category</strong></summary>\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "| Category | Size | Percentage |\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "|----------|-----:|----------:|\n"); err != nil {
		return err
	}

	for _, item := range sorted {
		percentage := float64(item.size) / float64(total) * 100
		if _, err := fmt.Fprintf(w, "| %s | %s | %.1f%% |\n",
			item.name, util.FormatBytes(item.size), percentage); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "\n</details>\n\n"); err != nil {
		return err
	}

	return nil
}

// writeLargestFiles writes the largest files section
func (f *MarkdownFormatter) writeLargestFiles(w io.Writer, report *types.Report) error {
	if len(report.LargestFiles) == 0 {
		return nil
	}

	// Limit to top 10
	limit := 10
	if len(report.LargestFiles) < limit {
		limit = len(report.LargestFiles)
	}

	// Calculate combined size
	var combinedSize int64
	for i := 0; i < limit; i++ {
		combinedSize += report.LargestFiles[i].Size
	}

	total := calculateUncompressedSize(&report.SizeBreakdown)

	if _, err := fmt.Fprintf(w, "<details>\n<summary><strong>📦 Top %d Largest Files</strong> (%s combined)</summary>\n\n",
		limit, util.FormatBytes(combinedSize)); err != nil {
		return err
	}

	for i := 0; i < limit; i++ {
		file := report.LargestFiles[i]
		percentage := float64(file.Size) / float64(total) * 100
		truncated := truncatePath(file.Path, 80)
		if _, err := fmt.Fprintf(w, "%d. `%s` - %s (%.1f%%)\n",
			i+1, truncated, util.FormatBytes(file.Size), percentage); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "\n</details>\n\n"); err != nil {
		return err
	}

	return nil
}

// writeDuplicates writes the duplicate files section
func (f *MarkdownFormatter) writeDuplicates(w io.Writer, report *types.Report) error {
	if len(report.Duplicates) == 0 {
		return nil
	}

	// Calculate total wasted space
	var totalWasted int64
	for _, dup := range report.Duplicates {
		totalWasted += dup.WastedSize
	}

	if _, err := fmt.Fprintf(w, "<details>\n<summary><strong>🔄 Duplicate Files Found</strong> (%d sets, %s wasted)</summary>\n\n",
		len(report.Duplicates), util.FormatBytes(totalWasted)); err != nil {
		return err
	}

	for i, dup := range report.Duplicates {
		if i > 0 {
			if _, err := fmt.Fprintf(w, "---\n\n"); err != nil {
				return err
			}
		}

		fileName := dup.Files[0]
		if idx := strings.LastIndex(fileName, "/"); idx >= 0 {
			fileName = fileName[idx+1:]
		}

		if _, err := fmt.Fprintf(w, "#### %d copies of `%s` (%s each)\n",
			dup.Count, fileName, util.FormatBytes(dup.Size)); err != nil {
			return err
		}

		for _, path := range dup.Files {
			truncated := truncatePath(path, 80)
			if _, err := fmt.Fprintf(w, "- `%s`\n", truncated); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintf(w, "- **Wasted:** %s\n\n", util.FormatBytes(dup.WastedSize)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "**Total Wasted Space:** %s\n\n", util.FormatBytes(totalWasted)); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "</details>\n\n"); err != nil {
		return err
	}

	return nil
}

// writeOptimizations writes optimization section for a specific category
func (f *MarkdownFormatter) writeOptimizations(
	w io.Writer,
	opts []types.Optimization,
	categoryName string,
	emoji string,
	open bool,
) error {
	openAttr := ""
	if open {
		openAttr = " open"
	}

	totalSavings := calculateSavings(opts)

	if _, err := fmt.Fprintf(w, "<details%s>\n<summary><strong>%s %s</strong>",
		openAttr, emoji, categoryName); err != nil {
		return err
	}

	if len(opts) > 0 {
		if _, err := fmt.Fprintf(w, " (%d issues, %s savings)", len(opts), util.FormatBytes(totalSavings)); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "</summary>\n\n"); err != nil {
		return err
	}

	if len(opts) == 0 {
		if _, err := fmt.Fprintf(w, "✅ No issues found!\n\n"); err != nil {
			return err
		}
	} else {
		for i, opt := range opts {
			if i > 0 {
				if _, err := fmt.Fprintf(w, "---\n\n"); err != nil {
					return err
				}
			}

			if _, err := fmt.Fprintf(w, "#### %s\n", opt.Title); err != nil {
				return err
			}

			if opt.Description != "" {
				if _, err := fmt.Fprintf(w, "%s\n", opt.Description); err != nil {
					return err
				}
			}

			if _, err := fmt.Fprintf(w, "- **Impact:** %s\n", util.FormatBytes(opt.Impact)); err != nil {
				return err
			}

			if opt.Action != "" {
				if _, err := fmt.Fprintf(w, "- **Action:** %s\n", opt.Action); err != nil {
					return err
				}
			}

			if len(opt.Files) > 0 {
				if _, err := fmt.Fprintf(w, "\n**Files:**\n"); err != nil {
					return err
				}
				for _, file := range opt.Files {
					truncated := truncatePath(file, 80)
					if _, err := fmt.Fprintf(w, "- `%s`\n", truncated); err != nil {
						return err
					}
				}
			}

			if _, err := fmt.Fprintf(w, "\n"); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintf(w, "</details>\n\n"); err != nil {
		return err
	}

	return nil
}

// writeByExtension writes the size by extension section
func (f *MarkdownFormatter) writeByExtension(w io.Writer, report *types.Report) error {
	if len(report.SizeBreakdown.ByExtension) == 0 {
		return nil
	}

	sorted := sortBySize(report.SizeBreakdown.ByExtension)
	total := calculateUncompressedSize(&report.SizeBreakdown)

	if _, err := fmt.Fprintf(w, "<details>\n<summary><strong>🔍 Size by File Extension</strong></summary>\n\n"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "| Extension | Size | Percentage |\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "|-----------|-----:|----------:|\n"); err != nil {
		return err
	}

	for _, item := range sorted {
		percentage := float64(item.size) / float64(total) * 100
		if _, err := fmt.Fprintf(w, "| %s | %s | %.1f%% |\n",
			item.name, util.FormatBytes(item.size), percentage); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "\n</details>\n\n"); err != nil {
		return err
	}

	return nil
}

// Helper functions

// truncatePath truncates long paths for readability
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	// Truncate in middle: "very/long/.../path/file.ext"
	partLen := (maxLen - 5) / 2 // 5 for ".../"
	if partLen <= 0 {
		partLen = 1
	}
	return path[:partLen] + "/.../" + path[len(path)-partLen:]
}

// getCategoryGroups groups optimizations by category
func getCategoryGroups(opts []types.Optimization) map[string][]types.Optimization {
	groups := make(map[string][]types.Optimization)
	for _, opt := range opts {
		groups[opt.Category] = append(groups[opt.Category], opt)
	}
	return groups
}

// calculateSavings sums impact from optimizations
func calculateSavings(opts []types.Optimization) int64 {
	var total int64
	for _, opt := range opts {
		total += opt.Impact
	}
	return total
}

type sortedItem struct {
	name string
	size int64
}

// sortBySize sorts a map by value descending
func sortBySize(breakdown map[string]int64) []sortedItem {
	var sorted []sortedItem
	for k, v := range breakdown {
		sorted = append(sorted, sortedItem{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].size > sorted[j].size
	})

	return sorted
}

// countFiles counts total files in file tree
func countFiles(nodes []*types.FileNode) int {
	count := 0
	for _, node := range nodes {
		if !node.IsDir {
			count++
		}
		count += countFiles(node.Children)
	}
	return count
}

// calculateUncompressedSize sums all category sizes
func calculateUncompressedSize(breakdown *types.SizeBreakdown) int64 {
	return breakdown.Executable + breakdown.Frameworks +
		breakdown.Resources + breakdown.Assets +
		breakdown.Libraries + breakdown.DEX +
		breakdown.Other
}

// findLargestCategory finds the category with the largest size
func findLargestCategory(breakdown *types.SizeBreakdown) (string, int64) {
	categories := map[string]int64{
		"Frameworks": breakdown.Frameworks,
		"Resources":  breakdown.Resources,
		"Executable": breakdown.Executable,
		"Assets":     breakdown.Assets,
		"Libraries":  breakdown.Libraries,
		"DEX":        breakdown.DEX,
		"Other":      breakdown.Other,
	}

	var maxName string
	var maxSize int64
	for name, size := range categories {
		if size > maxSize {
			maxSize = size
			maxName = name
		}
	}

	return maxName, maxSize
}
