package filestore

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/utils"
)

const pathRegexPattern = `^[/](` + utils.NameRegexFragment + `[\/]?)*$`

var pathRegexExp = regexp.MustCompile(pathRegexPattern)

// SanitizePath standardizes and sanitizes the path, and validates it against naming requirements.
func SanitizePath(path string) (string, error) {
	path = filepath.Clean(filepath.Join("/", strings.TrimPrefix(path, "/")))
	cleanedPath := filepath.Clean(path) // filepath.Clean() is unnecessary and can lead to potential issues,
	// especially when dealing with URLs or paths containing dot-segments (e.g., /../ or /./).
	if !pathRegexExp.MatchString(path) {
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
