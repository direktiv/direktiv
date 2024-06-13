package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

func findProjectRoot(path string) (string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !fi.IsDir() {
		path = filepath.Dir(path)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for i := range files {
		file := files[i]
		if file.Name() == ".direktivignore" {
			return path, nil
		}
	}

	newPath := filepath.Dir(path)
	if path == newPath {
		return "", fmt.Errorf("no .direktivignore file found")
	}

	return findProjectRoot(newPath)
}
