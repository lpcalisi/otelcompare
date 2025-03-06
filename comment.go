package main

import (
	"fmt"
	"io/ioutil"

	"github.com/lcalisi/otel-html-viewer/pkg/github"
	"github.com/lcalisi/otel-html-viewer/pkg/trace"
	"github.com/spf13/cobra"
)

var (
	commentInputFile string
	commentPrNumber  int
	commentToken     string
	commentOwner     string
	commentRepo      string
)

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Comment traces on a GitHub PR",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runComment(commentInputFile)
	},
}

func init() {
	commentCmd.Flags().StringVarP(&commentInputFile, "input", "i", "", "Input JSON file containing traces")
	commentCmd.Flags().IntVarP(&commentPrNumber, "pr", "p", 0, "Pull request number to comment on")
	commentCmd.Flags().StringVarP(&commentToken, "token", "t", "", "GitHub token for PR comments")
	commentCmd.Flags().StringVar(&commentOwner, "owner", "", "GitHub repository owner")
	commentCmd.Flags().StringVar(&commentRepo, "repo", "", "GitHub repository name")

	commentCmd.MarkFlagRequired("input")
	commentCmd.MarkFlagRequired("pr")
	commentCmd.MarkFlagRequired("token")
	commentCmd.MarkFlagRequired("owner")
	commentCmd.MarkFlagRequired("repo")

	rootCmd.AddCommand(commentCmd)
}

func runComment(inputFile string) error {
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
	comment := fmt.Sprintf("### OpenTelemetry Traces Comparison\n\n%s", markdown)

	// Comment on the PR
	client := github.NewClient(commentToken)
	if err := client.CommentPR(commentOwner, commentRepo, commentPrNumber, comment); err != nil {
		return fmt.Errorf("error commenting on PR: %w", err)
	}

	return nil
}
