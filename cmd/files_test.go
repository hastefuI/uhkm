package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCollectUHKMFilesSorted(t *testing.T) {
	dir := t.TempDir()
	want := []string{
		filepath.Join(dir, "a.uhkm"),
		filepath.Join(dir, "b.uhkm"),
		filepath.Join(dir, "nested", "c.uhkm"),
	}

	for _, path := range want {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "ignore.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := collectUHKMFiles([]string{dir})
	if err != nil {
		t.Fatalf("collectUHKMFiles: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCollectUHKMFilesDeduplicatesOverlappingInputs(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.uhkm")
	if err := os.WriteFile(file, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := collectUHKMFiles([]string{dir, file})
	if err != nil {
		t.Fatalf("collectUHKMFiles: %v", err)
	}

	want := []string{file}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
