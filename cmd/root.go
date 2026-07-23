// Package cmd implements the anvil command line.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "anvil",
	Short: "Guided, zero-config build and release pipeline for mobile and app projects",
	Long: `anvil detects a project's stack and runs the right build and release
lifecycle with one guided command, without memorizing each framework's CLI.

anvil is under active development. See docs/ROADMAP.md.`,
	SilenceUsage: true,
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
