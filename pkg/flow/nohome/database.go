package nohome

import (
	"strings"
)

type HasAttributes interface {
	GetAttributes() map[string]string
}

func GetWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
