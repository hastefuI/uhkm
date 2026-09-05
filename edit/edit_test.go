package edit_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.hasteful.org/uhkm/edit"
)

func TestWriteCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.uhkm")
	data := []byte("hello\n")

	if err := edit.Write(path, data); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}
}

func TestWriteOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.uhkm")

	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	newData := []byte("new content\n")
	if err := edit.Write(path, newData); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(newData) {
		t.Errorf("got %q, want %q", got, newData)
	}
}

func TestWriteLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.uhkm")

	if err := edit.Write(path, []byte("data\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("expected 1 file, found %v", names)
	}
}
