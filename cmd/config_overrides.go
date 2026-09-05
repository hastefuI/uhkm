package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.hasteful.org/uhkm/config"
)

var (
	overrideIndentStyle      string
	overrideIndentWidth      int
	overrideNamingConvention string
)

func addConfigOverrideFlags(command *cobra.Command) {
	command.Flags().StringVar(&overrideIndentStyle, "indent-style", "", `override indentation style ("spaces" or "tabs")`)
	command.Flags().IntVar(&overrideIndentWidth, "indent-width", 0, "override indentation width when style is spaces")
	command.Flags().StringVar(&overrideNamingConvention, "naming-convention", "", `override naming convention ("kebab", "snake", or "pascal")`)
}

func applyConfigOverrides(command *cobra.Command, cfg *config.Config) error {
	if command.Flags().Changed("indent-style") {
		cfg.Lint.Indentation.Style = overrideIndentStyle
	}
	if command.Flags().Changed("indent-width") {
		cfg.Lint.Indentation.Width = overrideIndentWidth
	}
	if command.Flags().Changed("naming-convention") {
		cfg.Lint.Naming.Convention = overrideNamingConvention
	}

	config.Normalize(cfg)
	if err := config.Validate(*cfg); err != nil {
		return fmt.Errorf("invalid effective config: %w", err)
	}
	return nil
}

func init() {
	addConfigOverrideFlags(checkCmd)
	addConfigOverrideFlags(formatCmd)
	addConfigOverrideFlags(configCmd)
}
