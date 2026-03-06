package discovery

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// GoFiles walks the directory tree rooted at root and returns a sorted list
// of all .go file paths found. The following directories are skipped entirely:
// .git, vendor, and node_modules. Returned paths are absolute (or relative to
// the process working directory, matching root's own form). An error is
// returned only when the walk itself fails; a directory that simply contains
// no Go files returns an empty slice.
func GoFiles(root string) ([]string, error) {
	paths := make([]string, 0, 128)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(paths)
	return paths, nil
}
