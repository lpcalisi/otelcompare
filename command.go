package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "otel-html-viewer",
	Short: "Generate and compare OpenTelemetry traces",
	Long: `A tool that reads JSON files with OpenTelemetry traces,
generates visualizations and compares them in GitHub Pull Requests.`,
}

func Execute() error {
	return rootCmd.Execute()
}
