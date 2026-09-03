package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

// validPreamble is a spec-conformant header for the "macro.uhkm" fixtures, so
// that CLI tests exercise the behavior under test rather than tripping the
// filename-comment or required-pragma rules.
const validPreamble = "// macro.uhkm\n// @uhkm-name: macro\n// @uhkm-version: 1.0.0\n\n"

func resetCommandState() {
	fixFlag = false
	overrideIndentStyle = ""
	overrideIndentWidth = 0
	overrideNamingConvention = ""

	root.SetArgs(nil)
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	root.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
	root.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
	checkCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
	formatCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
	configCmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = flag.Value.Set(flag.DefValue)
		flag.Changed = false
	})
}

func executeWithArgs(args ...string) error {
	resetCommandState()
	root.SetArgs(args)
	return Execute()
}

func TestCheckFixAppliesFormatting(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "macro.uhkm")
	if err := os.WriteFile(file, []byte(validPreamble+"action {\n\tkey A\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := executeWithArgs("check", "--fix", file); err != nil {
		t.Fatalf("check --fix failed: %v", err)
	}

	got, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	want := validPreamble + "action {\n    key A\n}\n"
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCheckReturnsLintFailure(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "macro.uhkm")
	if err := os.WriteFile(file, []byte(validPreamble+"action {\n\tkey A\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := executeWithArgs("check", file)
	if !errors.Is(err, ErrLintFailure) {
		t.Fatalf("err = %v, want ErrLintFailure", err)
	}
}

func TestCLIFlagsOverrideResolvedConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := `[lint.indentation]
style = "tabs"
`
	if err := os.WriteFile(filepath.Join(dir, ".uhkm.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(dir, "macro.uhkm")
	if err := os.WriteFile(file, []byte(validPreamble+"action {\n\tkey A\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := executeWithArgs("check", file); err != nil {
		t.Fatalf("check with file config failed: %v", err)
	}

	err := executeWithArgs("check", "--indent-style", "spaces", file)
	if !errors.Is(err, ErrLintFailure) {
		t.Fatalf("err = %v, want ErrLintFailure when CLI override wins", err)
	}
}
