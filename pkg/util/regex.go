package util

import (
	"log"
	"regexp"
)

const RegexPattern = `^(([a-zA-Z][\w\-]*[a-zA-Z0-9])|([a-zA-Z]))$`

var reg *regexp.Regexp

func init() {

	var err error
	reg, err = regexp.Compile(RegexPattern)
	if err != nil {
		log.Fatal(err.Error())
	}

}

// MatchesRegex responds true if the provided string matches the
// RegexPattern constant defined in this package.
func MatchesRegex(s string) bool {
	return reg.MatchString(s)
}
