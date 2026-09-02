package cmd

import (
	"bytes"

	"github.com/spf13/cobra"

	"github.com/hastefuI/uhkm/config"
	"github.com/hastefuI/uhkm/edit"
	"github.com/hastefuI/uhkm/format"
)

var formatCmd = &cobra.Command{
	Use:   "format [paths...]",
	Short: "Format .uhkm files in place",
	Long: `Format .uhkm files in place.

Canonicalizes pragmas, strips trailing whitespace, normalizes indentation,
and ensures exactly one blank line between the preamble and the body.
Defaults to the current directory when no paths are provided.

Exit codes:
  0  success
  2  tool error`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return forEachResolvedFile(cmd, args, func(path string, cfg config.Config, content []byte) error {
			formatted := format.Format(content, cfg)
			if bytes.Equal(content, formatted) {
				return nil // already canonical; skip the write
			}

			if err := edit.Write(path, formatted); err != nil {
				return err
			}
			return nil
		})
	},
}
