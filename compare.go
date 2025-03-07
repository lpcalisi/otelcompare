package main

import (
	"fmt"
	"os"

	"github.com/lcalisi/otel-html-viewer/pkg/github"
	"github.com/lcalisi/otel-html-viewer/pkg/trace"
	"github.com/spf13/cobra"
)

var (
	compareInputFiles []string
	comparePrNumber   int
	compareToken      string
	compareOwner      string
	compareRepo       string
	compareAttribute  string
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare traces between different files",
	Long: `Compare traces between different files and generate a markdown report.
For example:
  otel-html-viewer compare -i file1.json -i file2.json -i file3.json
  otel-html-viewer compare -i file1.json -i file2.json -a http.url`,
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

		// Check if we should comment on GitHub
		if compareToken != "" && comparePrNumber > 0 && compareOwner != "" && compareRepo != "" {
			client := github.NewClient(compareToken)
			return client.CommentPR(compareOwner, compareRepo, comparePrNumber, markdown)
		}

		// Otherwise, print to stdout
		fmt.Print(markdown)
		return nil
	},
}

func init() {
	compareCmd.Flags().StringArrayVarP(&compareInputFiles, "input", "i", []string{}, "Input JSON files to compare")
	compareCmd.Flags().IntVarP(&comparePrNumber, "pr", "p", 0, "Pull request number to comment on")
	compareCmd.Flags().StringVarP(&compareToken, "token", "t", "", "GitHub token for PR comments")
	compareCmd.Flags().StringVar(&compareOwner, "owner", "", "GitHub repository owner")
	compareCmd.Flags().StringVar(&compareRepo, "repo", "", "GitHub repository name")
	compareCmd.Flags().StringVarP(&compareAttribute, "attribute", "a", "name", "Attribute to use for trace identification (default: span name)")

	compareCmd.MarkFlagRequired("input")
	compareCmd.MarkFlagRequired("token")
	compareCmd.MarkFlagRequired("owner")
	compareCmd.MarkFlagRequired("repo")

	rootCmd.AddCommand(compareCmd)
}
