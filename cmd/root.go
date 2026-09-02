// Package cmd wires together the uhkm CLI commands using Cobra.
package cmd

import "github.com/spf13/cobra"

var root = &cobra.Command{
	Use:          "uhkm",
	Short:        "Lint and format Ultimate Hacking Keyboard Macro (.uhkm) files",
	SilenceUsage: true,
	// Errors are printed by main, not by Cobra.
	SilenceErrors: true,
}

// Execute runs the root command and returns any error.
func Execute() error {
	return root.Execute()
}

func init() {
	root.AddCommand(checkCmd)
	root.AddCommand(formatCmd)
	root.AddCommand(configCmd)
	root.AddCommand(versionCmd)
}
