package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.hasteful.org/uhkm/config"
)

func forEachResolvedFile(command *cobra.Command, paths []string, visit func(path string, cfg config.Config, content []byte) error) error {
	files, err := collectUHKMFiles(paths)
	if err != nil {
		return err
	}

	for _, path := range files {
		cfg, err := config.Resolve(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := applyConfigOverrides(command, &cfg); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		if err := visit(path, cfg, content); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	return nil
}
