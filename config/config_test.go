package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.hasteful.org/uhkm/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.Lint.Indentation.Style != "spaces" {
		t.Errorf("default style = %q, want %q", cfg.Lint.Indentation.Style, "spaces")
	}
	if cfg.Lint.Indentation.Width != 4 {
		t.Errorf("default width = %d, want 4", cfg.Lint.Indentation.Width)
	}
	if cfg.Lint.Naming.Convention != "kebab" {
		t.Errorf("default convention = %q, want %q", cfg.Lint.Naming.Convention, "kebab")
	}
}

func TestResolveNoConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Resolve(filepath.Join(dir, "my-macro.uhkm"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := config.Default()
	if cfg != want {
		t.Errorf("config = %+v, want default %+v", cfg, want)
	}
}

func TestResolveFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgContent := `[lint.indentation]
style = "tabs"
width = 0

[lint.naming]
convention = "snake"
`
	if err := os.WriteFile(filepath.Join(dir, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Resolve(filepath.Join(dir, "my-macro.uhkm"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Lint.Indentation.Style != "tabs" {
		t.Errorf("style = %q, want %q", cfg.Lint.Indentation.Style, "tabs")
	}
	if cfg.Lint.Naming.Convention != "snake" {
		t.Errorf("convention = %q, want %q", cfg.Lint.Naming.Convention, "snake")
	}
}

func TestResolveStopsAtGitBoundary(t *testing.T) {
	root := t.TempDir()

	// Place a non-default config in root.
	cfgContent := `[lint.indentation]
style = "tabs"
`
	if err := os.WriteFile(filepath.Join(root, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a sub-project with its own .git and no .uhkm.toml.
	sub := filepath.Join(root, "sub")
	if err := os.MkdirAll(filepath.Join(sub, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Resolving from inside sub should not find root/.uhkm.toml.
	cfg, err := config.Resolve(filepath.Join(sub, "file.uhkm"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Lint.Indentation.Style != "spaces" {
		t.Errorf("expected default style after .git boundary, got %q", cfg.Lint.Indentation.Style)
	}
}

func TestResolveSearchesUpward(t *testing.T) {
	root := t.TempDir()

	cfgContent := `[lint.indentation]
style = "tabs"
`
	if err := os.WriteFile(filepath.Join(root, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a nested directory with no config of its own.
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Resolve(filepath.Join(nested, "file.uhkm"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Lint.Indentation.Style != "tabs" {
		t.Errorf("style = %q, want %q", cfg.Lint.Indentation.Style, "tabs")
	}
}

func TestResolveDirectory(t *testing.T) {
	dir := t.TempDir()
	cfgContent := `[lint.naming]
convention = "pascal"
`
	if err := os.WriteFile(filepath.Join(dir, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Lint.Naming.Convention != "pascal" {
		t.Errorf("convention = %q, want %q", cfg.Lint.Naming.Convention, "pascal")
	}
}

func TestResolveInvalidIndentationStyle(t *testing.T) {
	dir := t.TempDir()
	cfgContent := `[lint.indentation]
style = "invalid"
`
	if err := os.WriteFile(filepath.Join(dir, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Resolve(filepath.Join(dir, "file.uhkm")); err == nil {
		t.Fatal("expected error for invalid indentation style, got nil")
	}
}

func TestResolveInvalidWidthInSpacesMode(t *testing.T) {
	dir := t.TempDir()
	cfgContent := `[lint.indentation]
style = "spaces"
width = 0
`
	if err := os.WriteFile(filepath.Join(dir, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Resolve(filepath.Join(dir, "file.uhkm")); err == nil {
		t.Fatal("expected error for non-positive width in spaces mode, got nil")
	}
}

func TestResolveInvalidNamingConvention(t *testing.T) {
	dir := t.TempDir()
	cfgContent := `[lint.naming]
convention = "camel"
`
	if err := os.WriteFile(filepath.Join(dir, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := config.Resolve(filepath.Join(dir, "file.uhkm")); err == nil {
		t.Fatal("expected error for invalid naming convention, got nil")
	}
}

func TestResolveNormalizesKebabCaseAlias(t *testing.T) {
	dir := t.TempDir()
	cfgContent := `[lint.naming]
convention = "kebab-case"
`
	if err := os.WriteFile(filepath.Join(dir, ".uhkm.toml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Resolve(filepath.Join(dir, "file.uhkm"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Lint.Naming.Convention != "kebab" {
		t.Fatalf("convention = %q, want %q", cfg.Lint.Naming.Convention, "kebab")
	}
}
