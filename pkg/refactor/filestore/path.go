package filestore

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/util"
)

const pathRegexPattern = `^[/](` + util.NameRegexFragment + `[\/]?)*$`

var pathRegexExp = regexp.MustCompile(pathRegexPattern)

// SanitizePath standardizes and sanitized the path, and validates it against naming requirements.
func SanitizePath(path string) (string, error) {
	path = filepath.Join("/", path)
	path = filepath.Clean(path)

	if !pathRegexExp.MatchString(path) {
		return "", errors.New("invalid path string")
	}

	return path, nil
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
