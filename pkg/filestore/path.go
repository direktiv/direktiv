package filestore

import (
	"strings"
	"unicode/utf8"
)

// ValidatePath checks a POSIX-style absolute path with these rules:
// - must start with "/" (absolute)
// - "/" is allowed; otherwise no trailing slash
// - no empty segments (so no "//")
// - no "." or ".." segments anywhere
// It returns the input unchanged if valid.
func ValidatePath(path string) (string, error) {
	if path == "" || !utf8.ValidString(path) || strings.ContainsRune(path, 0) {
		return "", ErrInvalidPathParameter
	}
	if path == "/" {
		return path, nil
	}

	// Must start with a single "/"
	if !strings.HasPrefix(path, "/") {
		return "", ErrInvalidPathParameter
	}

	// "/" is the only allowed path that ends with "/"
	if path != "/" && strings.HasSuffix(path, "/") {
		return "", ErrInvalidPathParameter
	}

	// No double slashes anywhere
	if strings.Contains(path, "//") {
		return "", ErrInvalidPathParameter
	}

	// Check segments for "." or ".."
	// Split keeps a leading empty element for the first "/", so skip index 0.
	segs := strings.Split(path, "/")
	for i, s := range segs {
		if i == 0 {
			continue // leading empty segment before the first slash
		}
		if s == "" { // would indicate '//' or trailing '/', both already guarded
			return "", ErrInvalidPathParameter
		}
		if s == "." || s == ".." {
			return "", ErrInvalidPathParameter
		}
	}

	return path, nil
}

// GetPathDepth reads the path and returns the depth value.
func GetPathDepth(path string) int {
	depth := strings.Count(path, "/")
	if path == "/" {
		depth = 0
	}

	return depth
}
