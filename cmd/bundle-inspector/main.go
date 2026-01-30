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

func runAnalyze(cmd *cobra.Command, args []string) error {
	var artifactPath string

	// Determine artifact path
	if len(args) > 0 {
		// Explicit path provided
		artifactPath = args[0]
	} else if !noAutoDetect {
		// Try auto-detection from Bitrise environment
		detectedPath, err := bitrise.DetectBundlePath()
		if err != nil {
			return fmt.Errorf(
				"no artifact path provided and auto-detection failed: %w\n\n"+
					"Usage: bundle-inspector analyze <file-path>", err)
		}
		artifactPath = detectedPath
		fmt.Fprintf(os.Stderr, "Auto-detected artifact from Bitrise environment: %s\n", artifactPath)
	} else {
		return fmt.Errorf("no artifact path provided\n\nUsage: bundle-inspector analyze <file-path>")
	}

	// Validate file exists
	if _, err := os.Stat(artifactPath); err != nil {
		return fmt.Errorf("artifact not found: %w", err)
	}

	// Create orchestrator
	orch := orchestrator.New()
	orch.IncludeDuplicates = includeDuplicates

	// Perform analysis
	fmt.Fprintf(os.Stderr, "Analyzing %s...\n", artifactPath)
	if includeDuplicates {
		fmt.Fprintf(os.Stderr, "Detecting duplicates and additional optimizations...\n")
	}

	ctx := context.Background()
	analysisReport, err := orch.RunAnalysis(ctx, artifactPath)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Determine output destination
	var writer *os.File
	var actualOutputFile string

	if outputFile != "" {
		actualOutputFile = outputFile
	} else {
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
		actualOutputFile = fmt.Sprintf("bundle-analysis-%s.%s", artifactName, extension)
	}

	f, err := os.Create(actualOutputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()
	writer = f

	switch outputFormat {
	case "text":
		formatter := report.NewTextFormatter()
		if err := formatter.Format(writer, analysisReport); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	case "json":
		formatter := report.NewJSONFormatter(true)
		if err := formatter.Format(writer, analysisReport); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	case "markdown":
		formatter := report.NewMarkdownFormatter()
		if err := formatter.Format(writer, analysisReport); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	case "html":
		return fmt.Errorf("HTML output not yet implemented")
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	fmt.Fprintf(os.Stderr, "Report written to: %s\n", actualOutputFile)

	// Export to Bitrise deploy directory if in Bitrise environment
	if bitrise.IsBitriseEnvironment() {
		if err := exportToBitrise(analysisReport); err != nil {
			// Log warning but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to export to Bitrise deploy directory: %v\n", err)
		}
	}

	return nil
}

// exportToBitrise exports analysis results to Bitrise deploy directory
func exportToBitrise(analysisReport *types.Report) error {
	metadata := bitrise.GetBuildMetadata()
	if metadata.DeployDir == "" {
		return nil // Not in Bitrise or BITRISE_DEPLOY_DIR not set
	}

	// Export JSON report
	jsonData, err := json.MarshalIndent(analysisReport, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	jsonPath, err := bitrise.WriteToDeployDir("bundle-analysis.json", jsonData)
	if err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Exported JSON report to: %s\n", jsonPath)

	// Export text report
	textFormatter := report.NewTextFormatter()
	tmpFile, err := os.CreateTemp("", "bundle-report-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := textFormatter.Format(tmpFile, analysisReport); err != nil {
		return fmt.Errorf("failed to format text report: %w", err)
	}
	tmpFile.Close()

	textPath, err := bitrise.ExportToDeployDir(tmpFile.Name(), "bundle-report.txt")
	if err != nil {
		return fmt.Errorf("failed to export text report: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Exported text report to: %s\n", textPath)

	// Export markdown report (for easy GitHub/GitLab integration)
	markdownFormatter := report.NewMarkdownFormatter()
	tmpMdFile, err := os.CreateTemp("", "bundle-report-*.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create markdown temp file: %v\n", err)
	} else {
		defer os.Remove(tmpMdFile.Name())

		if err := markdownFormatter.Format(tmpMdFile, analysisReport); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to format markdown report: %v\n", err)
		} else {
			tmpMdFile.Close()

			mdPath, err := bitrise.ExportToDeployDir(tmpMdFile.Name(), "bundle-analysis.md")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to export markdown report: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "✓ Exported markdown report to: %s\n", mdPath)
			}
		}
	}

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
