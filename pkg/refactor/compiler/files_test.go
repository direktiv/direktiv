package compiler_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
)

func TestFiles(t *testing.T) {

	def := `
	const fileOne = getFile({
		name: "jens/myfile.txt",
		permission: 755,
		scope: "shared",
	});	  
	`

	c, _ := compiler.New("", def)
	c.CompileFlow()
}
