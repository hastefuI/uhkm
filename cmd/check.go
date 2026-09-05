package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.hasteful.org/uhkm/config"
	"go.hasteful.org/uhkm/edit"
	"go.hasteful.org/uhkm/lint"
)

var fixFlag bool

var checkCmd = &cobra.Command{
	Use:   "check [paths...]",
	Short: "Run lint checks on .uhkm files",
	Long: `Run lint checks on .uhkm files.

Applies UHKM100 (indentation), UHKM200 to UHKM201 (file naming), UHKM300 to
UHKM305 (preamble pragmas) and UHKM400 (file encoding).
Defaults to the current directory when no paths are provided.

Exit codes:
  0  no issues found
  1  lint issues found
  2  tool error`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var hasIssues bool
		if err := forEachResolvedFile(cmd, args, func(path string, cfg config.Config, content []byte) error {
			if fixFlag {
				fixed, changed := lint.Fix(content, cfg)
				if changed {
					if err := edit.Write(path, fixed); err != nil {
						return err
					}
					content = fixed
				}
			}

			for _, iss := range lint.Check(path, content, cfg) {
				fmt.Fprintln(cmd.ErrOrStderr(), iss)
				hasIssues = true
			}
			return nil
		}); err != nil {
			return err
		}

		if hasIssues {
			return ErrLintFailure
		}
		return nil
	},
}

func init() {
	checkCmd.Flags().BoolVar(&fixFlag, "fix", false, "automatically fix issues where possible")
}
