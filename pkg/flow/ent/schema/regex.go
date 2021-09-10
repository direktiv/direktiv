package schema

import "regexp"

const NameRegexFragment = `(([a-z][a-z0-9_\-]*[a-z0-9])|([a-z]))`

const NameRegexPattern = `^` + NameRegexFragment + `$`

var NameRegex = regexp.MustCompile(NameRegexPattern)

const URIRegexPattern = `^(` + NameRegexFragment + `[\/]?)*$`

var URIRegex = regexp.MustCompile(URIRegexPattern)

const VarNameRegexPattern = `(([a-zA-Z][a-zA-Z0-9_]*[a-zA-Z0-9])|([a-zA-Z]))`

var VarNameRegex = regexp.MustCompile(VarNameRegexPattern)
