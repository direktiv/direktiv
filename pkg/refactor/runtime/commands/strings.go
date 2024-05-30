package commands

import "strings"

func Trim(in string) string {
	return strings.TrimSuffix(in, "\n")
}
