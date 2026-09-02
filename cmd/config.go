package cmd

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	"github.com/hastefuI/uhkm/config"
)

var configCmd = &cobra.Command{
	Use:   "config [paths...]",
	Short: "Print resolved configuration",
	Long: `Print the effective configuration for each path.

Searches upward from each path for a .uhkm.toml file, stopping at a .git/
boundary. Prints defaults when no config file is found.
Defaults to the current directory when no paths are provided.

Exit codes:
  0  success
  2  config error`,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := args
		if len(paths) == 0 {
			paths = []string{"."}
		}

		for i, path := range paths {
			cfg, err := config.Resolve(path)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if err := applyConfigOverrides(cmd, &cfg); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}

			if len(paths) > 1 {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("# %s\n", path)
			}

			enc := toml.NewEncoder(os.Stdout)
			if err := enc.Encode(cfg); err != nil {
				return err
			}
		}
		return nil
	},
}
