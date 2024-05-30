package util

import (
	"log"
	"regexp"
)

const NameRegexFragment = `(([a-zA-Z][a-zA-Z0-9_\-\.]*[a-zA-Z0-9])|([a-zA-Z]))`

const NameRegexPattern = `^` + NameRegexFragment + `$`

const VarNameRegexPattern = `^(([a-zA-Z][a-zA-Z0-9_\-\.]*[a-zA-Z0-9])|([a-zA-Z]))$`

const VarSecretNameAndSecretsFolderNamePattern = `^(([a-zA-Z][a-zA-Z0-9_\-\./]*[a-zA-Z0-9/])|([a-zA-Z/]))$`

var VarNameRegex = regexp.MustCompile(VarNameRegexPattern)

const (
	RegexPattern    = NameRegexPattern
	VarRegexPattern = VarNameRegexPattern
)

var (
	reg    *regexp.Regexp
	varreg *regexp.Regexp
)

func init() {
	var err error
	reg, err = regexp.Compile(RegexPattern)
	if err != nil {
		log.Fatal(err.Error())
	}

	varreg, err = regexp.Compile(VarRegexPattern)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = regexp.Compile(VarSecretNameAndSecretsFolderNamePattern)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// MatchesRegex responds true if the provided string matches the
// RegexPattern constant defined in this package.
func MatchesRegex(s string) bool {
	return reg.MatchString(s)
}

func MatchesVarRegex(s string) bool {
	return varreg.MatchString(s)
}
