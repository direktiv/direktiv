package compiler_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
)

func TestSecrets(t *testing.T) {

	def := `
	getSecret({
		name: "mysecret"
	})
	`

	c, _ := compiler.New("", def)
	c.CompileFlow()
}
