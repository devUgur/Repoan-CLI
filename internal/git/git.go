package git

import (
	"os"
	"path/filepath"
)

// FindRoot searches for a .git directory starting from the given path and moving up.
func FindRoot(start string) (string, error) {
	absStart, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	dir := absStart
	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .git
			return "", nil
		}
		dir = parent
	}
}
