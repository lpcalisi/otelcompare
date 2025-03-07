package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lpcalisi/otelcompare/pkg/github"
	"github.com/lpcalisi/otelcompare/pkg/trace"
	"github.com/spf13/cobra"
)

var (
	infoInputFile string
	infoPrNumber  int
	infoOwner     string
	infoRepo      string
	infoDryRun    bool
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Generate trace information for a GitHub PR",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInfo(infoInputFile)
	},
}

func init() {
	infoCmd.Flags().StringVarP(&infoInputFile, "input", "i", "", "Input JSON file containing traces")
	infoCmd.Flags().IntVarP(&infoPrNumber, "pr", "p", 0, "Pull request number to comment on")
	infoCmd.Flags().StringVar(&infoOwner, "owner", "", "GitHub repository owner")
	infoCmd.Flags().StringVar(&infoRepo, "repo", "", "GitHub repository name")
	infoCmd.Flags().BoolVar(&infoDryRun, "dry-run", false, "Print comment to stdout without posting to GitHub")

	infoCmd.MarkFlagRequired("input")
	infoCmd.MarkFlagRequired("pr")

	rootCmd.AddCommand(infoCmd)
}

func runInfo(inputFile string) error {
	// Read input file
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading input file: %w", err)
	}

	// Parse traces
	traces, err := trace.ParseTraces(data)
	if err != nil {
		return fmt.Errorf("error parsing traces: %w", err)
	}

	// Generate Markdown for the PR comment
	markdown := trace.GenerateMarkdown(traces)
	comment := fmt.Sprintf("### OpenTelemetry Traces Analysis\n\n%s", markdown)

	// If dry-run, just print to stdout
	if infoDryRun {
		fmt.Print(comment)
		return nil
	}

	// Validate GitHub flags if not dry-run
	if infoOwner == "" || infoRepo == "" {
		return fmt.Errorf("--owner and --repo are required when not using --dry-run")
	}

	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required when not using --dry-run")
	}

	// Comment on the PR
	client := github.NewClient(token)
	if err := client.CommentPR(infoOwner, infoRepo, infoPrNumber, comment); err != nil {
		return fmt.Errorf("error commenting on PR: %w", err)
	}

	return nil
}
