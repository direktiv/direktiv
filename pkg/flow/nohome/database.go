package nohome

import (
	"strings"

	"github.com/direktiv/direktiv/pkg/flow/nohome/recipient"
)

type HasAttributes interface {
	GetAttributes() map[string]string
}

func GetAttributes(recipientType recipient.RecipientType, a ...HasAttributes) map[string]string {
	m := make(map[string]string)
	m["recipientType"] = string(recipientType)
	for _, x := range a {
		y := x.GetAttributes()
		for k, v := range y {
			m[k] = v
		}
	}

	return m
}

func GetWorkflow(path string) string {
	return strings.Split(path, ":")[0]
}
