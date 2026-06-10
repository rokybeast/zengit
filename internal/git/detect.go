package git

import (
	"os"
	"path/filepath"
)

// walks up from dir looking for a .git folder
func IsRepo(dir string) bool {
	for {
		info, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil && info.IsDir() {
			return true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}
