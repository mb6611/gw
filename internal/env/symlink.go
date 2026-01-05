package env

import (
	"os"
	"path/filepath"
)

// DefaultPatterns are the files/directories to symlink from source to destination worktree
var DefaultPatterns = []string{
	".env",
	".env.*",
	".envrc",
	".claude",
}

// Symlink creates symlinks in dst for files matching patterns in src
// Skips files that don't exist in src
// Returns list of created symlinks
func Symlink(src, dst string, patterns []string) ([]string, error) {
	var created []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(src, pattern))
		if err != nil {
			// Invalid pattern, skip
			continue
		}

		for _, srcPath := range matches {
			name := filepath.Base(srcPath)
			dstPath := filepath.Join(dst, name)

			// Skip if already exists
			if _, err := os.Lstat(dstPath); err == nil {
				continue
			}

			// Create relative symlink for portability
			relPath, err := filepath.Rel(dst, srcPath)
			if err != nil {
				return created, err
			}

			if err := os.Symlink(relPath, dstPath); err != nil {
				return created, err
			}
			created = append(created, name)
		}
	}
	return created, nil
}
