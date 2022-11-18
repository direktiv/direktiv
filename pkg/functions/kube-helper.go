package functions

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	regex = "^[a-z]([-a-z0-9]{0,62}[a-z0-9])?$"
)

func SanitizeLabel(s string) string {
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSuffix(s, "/")
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "/", "-")

	if len(s) > 63 {
		s = s[:63]
	}

	return s
}

func validateLabel(name string) error {

	matched, err := regexp.MatchString(regex, name)
	if err != nil {
		return err
	}

	if !matched {
		return fmt.Errorf("invalid service name (must conform to regex: '%s')", regex)
	}

	return nil

}
