// Package edit provides atomic file write helpers used by the check --fix
// and format commands.
package edit

import (
	"os"
	"path/filepath"
)

// Write atomically replaces the file at path with data.
// It writes to a temporary file in the same directory, then renames it
// over the target to avoid partial writes.
func Write(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".uhkm-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
