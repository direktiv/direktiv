package filestore

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/util"
)

var pathRegexPattern = `^[/](` + util.NameRegexFragment + `[\/]?)*$`
var pathRegex = regexp.MustCompile(pathRegexPattern)

// SanitizePath standardizes and sanitized the path, and validates it against naming requirements.
func SanitizePath(path string) (string, error) {
	path = filepath.Clean(path)
	path = filepath.Join("/", path)
	if !pathRegex.MatchString(path) {
		return "", errors.New("path failed to match regex: " + pathRegexPattern)
	}
	return path, nil
}

// ParseDepth reads the path and returns the depth value. Use SanitizePath first, because if an error happens here the function may produce invalid results!
func ParseDepth(path string) int {
	depth := strings.Count(path, "/")
	if path == "/" {
		depth = 0
	}

	return depth
}
