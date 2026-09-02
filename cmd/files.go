package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// collectUHKMFiles expands paths into a list of .uhkm files.
// Directories are walked recursively. Files are passed through directly.
// If paths is empty, the current directory is used.
func collectUHKMFiles(paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(path) == ".uhkm" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	slices.Sort(files)
	files = slices.Compact(files)
	return files, nil
}
