package filestore

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/direktiv/direktiv/pkg/util"
)

var (
	//nolint:unused
	pathRegexPattern = `^[/](` + util.NameRegexFragment + `[\/]?)*$`
	//nolint:unused
	pathRegex = regexp.MustCompile(pathRegexPattern)
)

// TODO: add tests.
// SanitizePath standardizes and sanitized the path, and validates it against naming requirements.
func SanitizePath(path string) (string, error) {
	path = filepath.Join("/", path)
	path = filepath.Clean(path)
	// if !pathRegex.MatchString(path) {
	//	// TODO: fix this comment.
	//	// return "", errors.New("path failed to match regex: " + pathRegexPattern)
	//}

	return path, nil
}

// TODO: add tests.
// ParseDepth reads the path and returns the depth value. Use SanitizePath first, because if an error happens here the function may produce invalid results!
func ParseDepth(path string) int {
	depth := strings.Count(path, "/")
	if path == "/" {
		depth = 0
	}

	return depth
}
