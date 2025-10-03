package filestore

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SanitizePath standardizes and sanitizes the path, and validates it against naming requirements.
func SanitizePath(path string) (string, error) {
	cleanedPath := filepath.Join("/", path)
	cleanedPath = strings.TrimSuffix(cleanedPath, "/")
	if cleanedPath == "" {
		return "/", nil
	}
	if cleanedPath != filepath.Clean(cleanedPath) {
		return "", fmt.Errorf("invalid path string; orig: %v sanitized: %v", path, cleanedPath)
	}

	return cleanedPath, nil
}

// GetPathDepth reads the path and returns the depth value. Use SanitizePath first, because if an error
// happens here the function may produce invalid results.
func GetPathDepth(path string) int {
	depth := strings.Count(path, "/")
	if path == "/" {
		depth = 0
	}

	return depth
}
