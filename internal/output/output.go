package output

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir creates parent directories for the provided path when they do not exist.
func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	return nil
}
