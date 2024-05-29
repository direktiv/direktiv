package compiler_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
	"github.com/stretchr/testify/assert"
)

func TestFlowFunction(t *testing.T) {

	def := `
	var s = getSecret({ name: "ssss"})

	function start() {

	}

	var s = getSecret({ name: "ssss"})

	function stop() {

	}
	`

	c, _ := compiler.New("", def)
	fi, _ := c.CompileFlow()
	assert.Equal(t, "start", fi.Definition.State)

}
