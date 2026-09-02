package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Indentation holds UHKM100 rule settings.
type Indentation struct {
	Style string `toml:"style"` // "spaces" or "tabs"
	Width int    `toml:"width"` // indent width when Style == "spaces"
}

// Naming holds UHKM200 rule settings.
type Naming struct {
	Convention string `toml:"convention"` // "kebab", "snake", or "pascal"
}

// Lint groups all lint rule settings.
type Lint struct {
	Indentation Indentation `toml:"indentation"`
	Naming      Naming      `toml:"naming"`
}

// Config is the resolved configuration for a .uhkm file.
type Config struct {
	Lint Lint `toml:"lint"`
}

// Default returns the built-in default configuration.
func Default() Config {
	return Config{
		Lint: Lint{
			Indentation: Indentation{
				Style: "spaces",
				Width: 4,
			},
			Naming: Naming{
				Convention: "kebab",
			},
		},
	}
}

// Resolve loads the effective configuration for the given path.
// It searches upward from the path's directory for a .uhkm.toml file,
// stopping at a .git/ boundary or the filesystem root.
// Falls back to defaults if no config file is found.
func Resolve(path string) (Config, error) {
	cfg := Default()

	abs, err := filepath.Abs(path)
	if err != nil {
		return cfg, err
	}

	var dir string
	info, statErr := os.Stat(abs)
	if statErr == nil && info.IsDir() {
		dir = abs
	} else {
		dir = filepath.Dir(abs)
	}

	found, err := findConfig(dir)
	if err != nil || found == "" {
		return cfg, err
	}

	if _, err := toml.DecodeFile(found, &cfg); err != nil {
		return cfg, err
	}
	Normalize(&cfg)
	if err := Validate(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Normalize rewrites supported aliases to canonical configuration values.
func Normalize(cfg *Config) {
	if cfg.Lint.Naming.Convention == "kebab-case" {
		cfg.Lint.Naming.Convention = "kebab"
	}
}

// Validate enforces supported configuration values.
func Validate(cfg Config) error {
	switch cfg.Lint.Indentation.Style {
	case "spaces":
		if cfg.Lint.Indentation.Width <= 0 {
			return fmt.Errorf("lint.indentation.width must be > 0 when style is %q", cfg.Lint.Indentation.Style)
		}
	case "tabs":
		// width is ignored for tabs; allow any value to keep config minimal.
	default:
		return fmt.Errorf("lint.indentation.style must be %q or %q", "spaces", "tabs")
	}

	switch cfg.Lint.Naming.Convention {
	case "kebab", "snake", "pascal":
		return nil
	default:
		return fmt.Errorf("lint.naming.convention must be %q, %q, or %q", "kebab", "snake", "pascal")
	}
}

// findConfig walks upward from dir searching for .uhkm.toml.
// Returns "" if none is found before reaching .git/ or the fs root.
func findConfig(dir string) (string, error) {
	for {
		candidate := filepath.Join(dir, ".uhkm.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		// Stop at repository root.
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return "", nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}
