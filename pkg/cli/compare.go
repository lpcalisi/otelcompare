package cli

import (
	"fmt"
	"os"

	"github.com/lpcalisi/otelcompare/pkg/github"
	"github.com/lpcalisi/otelcompare/pkg/trace"
	"github.com/spf13/cobra"
)

var (
	compareInputFiles []string
	comparePrNumber   int
	compareOwner      string
	compareRepo       string
	compareAttribute  string
	compareDryRun     bool
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare traces between different files",
	Long: `Compare traces between different files and generate a markdown report.
For example:
  otelcompare compare -i file1.json -i file2.json -i file3.json
  otelcompare compare -i file1.json -i file2.json -a http.url`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(compareInputFiles) < 2 {
			return fmt.Errorf("at least two input files are required for comparison")
		}

		// Read and parse all files
		var traceSets []trace.TraceSet
		for _, file := range compareInputFiles {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("error reading file %s: %w", file, err)
			}

			traces, err := trace.ParseTraces(data)
			if err != nil {
				return fmt.Errorf("error parsing traces from %s: %w", file, err)
			}

			traceSets = append(traceSets, trace.TraceSet{
				Name:   file,
				Traces: traces,
			})
		}

		// Compare traces using the specified attribute
		markdown := trace.CompareMultipleTraces(traceSets, compareAttribute)

		// If dry-run, just print to stdout
		if compareDryRun {
			fmt.Print(markdown)
			return nil
		}

		// Validate GitHub flags if not dry-run
		if compareOwner == "" || compareRepo == "" {
			return fmt.Errorf("--owner and --repo are required when not using --dry-run")
		}

		// Get GitHub token from environment
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("GITHUB_TOKEN environment variable is required when not using --dry-run")
		}

		// Comment on GitHub
		client := github.NewClient(token)
		return client.CommentPR(compareOwner, compareRepo, comparePrNumber, markdown)
	},
}

func init() {
	compareCmd.Flags().StringArrayVarP(&compareInputFiles, "input", "i", []string{}, "Input JSON files to compare")
	compareCmd.Flags().IntVarP(&comparePrNumber, "pr", "p", 0, "Pull request number to comment on")
	compareCmd.Flags().StringVar(&compareOwner, "owner", "", "GitHub repository owner")
	compareCmd.Flags().StringVar(&compareRepo, "repo", "", "GitHub repository name")
	compareCmd.Flags().StringVarP(&compareAttribute, "attribute", "a", "trace_id", "Attribute to use for trace identification (default: span name)")
	compareCmd.Flags().BoolVar(&compareDryRun, "dry-run", false, "Print comment to stdout without posting to GitHub")

	compareCmd.MarkFlagRequired("input")

	rootCmd.AddCommand(compareCmd)
}
