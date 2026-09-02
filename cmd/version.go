package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These variables may be injected at build time via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("uhkm %s (commit %s, built %s)\n", version, commit, date)
	},
}
