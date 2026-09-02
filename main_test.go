package main

import (
	"os"
	"path/filepath"
	"testing"
)

func runWithArgs(t *testing.T, args ...string) int {
	t.Helper()
	previous := os.Args
	os.Args = append([]string{"uhkm"}, args...)
	t.Cleanup(func() {
		os.Args = previous
	})
	return run()
}

func TestRunExitCodeSuccess(t *testing.T) {
	if code := runWithArgs(t, "version"); code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
}

func TestRunExitCodeLintFailure(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "macro.uhkm")
	if err := os.WriteFile(file, []byte("action {\n\tkey A\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := runWithArgs(t, "check", "--indent-style", "spaces", file); code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
}

func TestRunExitCodeToolError(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "macro.uhkm")
	if err := os.WriteFile(file, []byte("action {\n    key A\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := runWithArgs(t, "check", "--indent-style", "invalid", file); code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
}
