package filestore

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ValidatePath(path string) (string, error) {
	cleanedPath := filepath.Join("/", path)
	cleanedPath = filepath.Clean(cleanedPath)
	cleanedPath = strings.TrimSuffix(cleanedPath, "/")
	if cleanedPath == "" {
		cleanedPath = "/"
	}
	if cleanedPath != path {
		return "", fmt.Errorf("invalid path string")
	}
	for _, s := range []string{"/./", "/../"} {
		if strings.Contains(cleanedPath, s) {
			return "", fmt.Errorf("invalid path string; contains '%v'", s)
		}
	}
	for _, s := range []string{".", "./", "/.", ""} {
		if cleanedPath == s {
			return "", fmt.Errorf("invalid path string")
		}
	}

	return cleanedPath, nil
}

// GetPathDepth reads the path and returns the depth value.
func GetPathDepth(path string) int {
	depth := strings.Count(path, "/")
	if path == "/" {
		depth = 0
	}

	return depth
}
