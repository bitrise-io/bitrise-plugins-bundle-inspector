// Bundle Inspector - Analyze mobile artifacts for size optimization opportunities
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/bitrise"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/orchestrator"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/internal/report"
	"github.com/bitrise-io/bitrise-plugins-bundle-inspector/pkg/types"
)

var (
	version = "0.1.1"
	commit  = "none"
	date    = "unknown"
)

var (
	outputFormats     string // Comma-separated list of formats
	outputFiles       string // Comma-separated list of filenames (optional)
	includeDuplicates bool
	noAutoDetect      bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "bundle-inspector",
	Short: "Analyze mobile artifacts for size optimization opportunities",
	Long: `Bundle Inspector analyzes mobile artifacts (iOS .ipa/.xcarchive/.app and
Android .apk/.aab) for size optimization opportunities. It detects duplicates,
bloat, and provides actionable recommendations.`,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file-path]",
	Short: "Analyze a mobile artifact",
	Long: `Analyze a mobile artifact (IPA, APK, AAB, App bundle, or XCArchive) and
generate a detailed size breakdown with optimization recommendations.

If no file path is provided, the tool will auto-detect the bundle path from
Bitrise environment variables (BITRISE_IPA_PATH, BITRISE_AAB_PATH, BITRISE_APK_PATH).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAnalyze,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Bundle Inspector %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built: %s\n", date)
	},
}

func init() {
	// Add commands
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(versionCmd)

	// Add flags
	analyzeCmd.Flags().StringVarP(&outputFormats, "output", "o", "text",
		"Output format(s) - comma-separated for multiple (text, json, markdown, html)")
	analyzeCmd.Flags().StringVarP(&outputFiles, "output-file", "f", "",
		"Output filename(s) - comma-separated when using multiple formats (default: auto-generated)")
	analyzeCmd.Flags().BoolVar(&includeDuplicates, "include-duplicates", true,
		"Enable duplicate file detection")
	analyzeCmd.Flags().BoolVar(&noAutoDetect, "no-auto-detect", false,
		"Disable auto-detection of bundle path from Bitrise environment")
}

// parseFormats parses and validates comma-separated output formats
func parseFormats(formatsStr string) ([]string, error) {
	formats := strings.Split(formatsStr, ",")
	validFormats := map[string]bool{
		"text":     true,
		"json":     true,
		"markdown": true,
		"html":     true,
	}

	var result []string
	seen := make(map[string]bool)

	for _, format := range formats {
		format = strings.TrimSpace(format)
		if format == "" {
			continue
		}

		// Check if valid
		if !validFormats[format] {
			return nil, fmt.Errorf("unsupported output format: %s (valid formats: text, json, markdown, html)", format)
		}

		// Check for duplicates
		if seen[format] {
			continue // Skip duplicates silently
		}

		seen[format] = true
		result = append(result, format)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid output formats specified")
	}

	return result, nil
}

// parseOutputFiles parses comma-separated output filenames
func parseOutputFiles(filesStr string) []string {
	if filesStr == "" {
		return nil
	}

	files := strings.Split(filesStr, ",")
	var result []string
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file != "" {
			result = append(result, file)
		}
	}
	return result
}

// getFileExtension returns the appropriate extension for a format
func getFileExtension(format string) string {
	switch format {
	case "json":
		return "json"
	case "markdown":
		return "md"
	case "text":
		return "txt"
	case "html":
		return "html"
	default:
		return "txt"
	}
}

// detectArtifactPath determines the artifact path from arguments or auto-detection
func detectArtifactPath(args []string) (string, error) {
	if len(args) > 0 {
		// Explicit path provided
		artifactPath := args[0]
		if _, err := os.Stat(artifactPath); err != nil {
			return "", fmt.Errorf("artifact not found: %w", err)
		}
		return artifactPath, nil
	}

	if noAutoDetect {
		return "", fmt.Errorf("no artifact path provided\n\nUsage: bundle-inspector analyze <file-path>")
	}

	// Try auto-detection from Bitrise environment
	detectedPath, err := bitrise.DetectBundlePath()
	if err != nil {
		return "", fmt.Errorf(
			"no artifact path provided and auto-detection failed: %w\n\n"+
				"Usage: bundle-inspector analyze <file-path>", err)
	}

	fmt.Fprintf(os.Stderr, "Auto-detected artifact from Bitrise environment: %s\n", detectedPath)
	return detectedPath, nil
}

// determineOutputFiles generates output filenames for all formats
func determineOutputFiles(artifactPath string, formats []string, explicitFiles []string) ([]string, error) {
	// If explicit filenames provided, validate count matches formats
	if len(explicitFiles) > 0 {
		if len(explicitFiles) != len(formats) {
			return nil, fmt.Errorf("number of output files (%d) must match number of formats (%d)", len(explicitFiles), len(formats))
		}
		return explicitFiles, nil
	}

	// Generate default filenames
	artifactName := filepath.Base(artifactPath)
	artifactName = strings.TrimSuffix(artifactName, filepath.Ext(artifactName))

	var filenames []string
	for _, format := range formats {
		extension := getFileExtension(format)
		filename := fmt.Sprintf("bundle-analysis-%s.%s", artifactName, extension)
		filenames = append(filenames, filename)
	}

	return filenames, nil
}

// writeReport writes the analysis report to a file using the specified format
func writeReport(filename string, format string, analysisReport *types.Report) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	switch format {
	case "text":
		formatter := report.NewTextFormatter()
		if err := formatter.Format(f, analysisReport); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	case "json":
		formatter := report.NewJSONFormatter(true)
		if err := formatter.Format(f, analysisReport); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	case "markdown":
		formatter := report.NewMarkdownFormatter()
		if err := formatter.Format(f, analysisReport); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	case "html":
		formatter := report.NewHTMLFormatter()
		if err := formatter.Format(f, analysisReport); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	return nil
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Determine artifact path
	artifactPath, err := detectArtifactPath(args)
	if err != nil {
		return err
	}

	// Parse and validate output formats
	formats, err := parseFormats(outputFormats)
	if err != nil {
		return err
	}

	// Parse explicit output filenames (if provided)
	explicitFiles := parseOutputFiles(outputFiles)

	// Generate output filenames
	filenames, err := determineOutputFiles(artifactPath, formats, explicitFiles)
	if err != nil {
		return err
	}

	// Create orchestrator and run analysis
	orch := orchestrator.New()
	orch.IncludeDuplicates = includeDuplicates

	fmt.Fprintf(os.Stderr, "Analyzing %s...\n", artifactPath)
	if includeDuplicates {
		fmt.Fprintf(os.Stderr, "Detecting duplicates and additional optimizations...\n")
	}

	ctx := context.Background()
	analysisReport, err := orch.RunAnalysis(ctx, artifactPath)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Write reports for all formats
	fmt.Fprintf(os.Stderr, "\nGenerating reports:\n")
	for i, format := range formats {
		filename := filenames[i]
		if err := writeReport(filename, format, analysisReport); err != nil {
			return fmt.Errorf("failed to write %s report: %w", format, err)
		}
		fmt.Fprintf(os.Stderr, "  ✓ %s: %s\n", strings.ToUpper(format), filename)
	}

	// Export to Bitrise deploy directory if in Bitrise environment
	if bitrise.IsBitriseEnvironment() {
		if err := exportToBitrise(analysisReport); err != nil {
			// Log warning but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to export to Bitrise deploy directory: %v\n", err)
		}
	}

	return nil
}

// exportJSONReport exports the report as JSON to the Bitrise deploy directory
func exportJSONReport(analysisReport *types.Report) error {
	jsonData, err := json.MarshalIndent(analysisReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	jsonPath, err := bitrise.WriteToDeployDir("bundle-analysis.json", jsonData)
	if err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Exported JSON report to: %s\n", jsonPath)
	return nil
}

// exportTextReport exports the report as text to the Bitrise deploy directory
func exportTextReport(analysisReport *types.Report) error {
	tmpFile, err := os.CreateTemp("", "bundle-report-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	formatter := report.NewTextFormatter()
	if err := formatter.Format(tmpFile, analysisReport); err != nil {
		return fmt.Errorf("failed to format text report: %w", err)
	}
	tmpFile.Close()

	exportPath, err := bitrise.ExportToDeployDir(tmpFile.Name(), "bundle-report.txt")
	if err != nil {
		return fmt.Errorf("failed to export text report: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Exported text report to: %s\n", exportPath)
	return nil
}

// exportMarkdownReport exports the report as markdown (best-effort, logs warnings on errors)
func exportMarkdownReport(analysisReport *types.Report) {
	tmpFile, err := os.CreateTemp("", "bundle-report-*.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create markdown temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	formatter := report.NewMarkdownFormatter()
	if err := formatter.Format(tmpFile, analysisReport); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to format markdown report: %v\n", err)
		return
	}
	tmpFile.Close()

	exportPath, err := bitrise.ExportToDeployDir(tmpFile.Name(), "bundle-analysis.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to export markdown report: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "✓ Exported markdown report to: %s\n", exportPath)
}

// exportToBitrise exports analysis results to Bitrise deploy directory
func exportToBitrise(analysisReport *types.Report) error {
	metadata := bitrise.GetBuildMetadata()
	if metadata.DeployDir == "" {
		return nil // Not in Bitrise or BITRISE_DEPLOY_DIR not set
	}

	// Export JSON report
	if err := exportJSONReport(analysisReport); err != nil {
		return err
	}

	// Export text report
	if err := exportTextReport(analysisReport); err != nil {
		return err
	}

	// Export markdown report (best-effort, don't fail on errors)
	exportMarkdownReport(analysisReport)

	// Log build metadata if available
	if metadata.BuildNumber != "" {
		fmt.Fprintf(os.Stderr, "  Build: #%s", metadata.BuildNumber)
		if metadata.CommitHash != "" {
			fmt.Fprintf(os.Stderr, " (%s)", metadata.CommitHash[:7])
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	return nil
}
