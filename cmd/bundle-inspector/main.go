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
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	outputFormat      string
	outputFile        string
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
	analyzeCmd.Flags().StringVarP(&outputFormat, "output", "o", "text",
		"Output format (text, json, markdown, html)")
	analyzeCmd.Flags().StringVarP(&outputFile, "output-file", "f", "",
		"Override default output filename (default: bundle-analysis-<artifact>.<format>)")
	analyzeCmd.Flags().BoolVar(&includeDuplicates, "include-duplicates", true,
		"Enable duplicate file detection")
	analyzeCmd.Flags().BoolVar(&noAutoDetect, "no-auto-detect", false,
		"Disable auto-detection of bundle path from Bitrise environment")
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

// determineOutputFile generates the output filename based on format and artifact
func determineOutputFile(artifactPath string) string {
	if outputFile != "" {
		return outputFile
	}

	// Extract artifact name from path
	artifactName := filepath.Base(artifactPath)
	// Remove extension
	artifactName = strings.TrimSuffix(artifactName, filepath.Ext(artifactName))

	// Generate default filename: bundle-analysis-<artifact>.<format>
	var extension string
	switch outputFormat {
	case "json":
		extension = "json"
	case "markdown":
		extension = "md"
	case "text":
		extension = "txt"
	case "html":
		extension = "html"
	default:
		extension = "txt"
	}
	return fmt.Sprintf("bundle-analysis-%s.%s", artifactName, extension)
}

// writeReport writes the analysis report to a file using the specified format
func writeReport(filename string, analysisReport *types.Report) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	switch outputFormat {
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
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Determine artifact path
	artifactPath, err := detectArtifactPath(args)
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

	// Write report to file
	outputFilename := determineOutputFile(artifactPath)
	if err := writeReport(outputFilename, analysisReport); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Report written to: %s\n", outputFilename)

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
