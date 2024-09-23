package helpers

import (
	"fmt"
	"path/filepath"
)

func ResolvePath(relativePath string) (string, error) {
	absPath, err := filepath.Abs(relativePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	return absPath, nil
}
