package compiler_test

import (
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/compiler"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDefinition(t *testing.T) {

	emptyDef := ``
	c, _ := compiler.New("", emptyDef)
	info, _ := c.CompileFlow()

	assert.True(t, info.Definition.Json)
	fmt.Printf("%+v\n", info.Definition)
}

func TestNoFunctionsOutside(t *testing.T) {

	def := `
	a()

	function a() {}

	function b() {}
	`
	_, err := compiler.New("", def)
	assert.Error(t, err)

	def = `
	var test = b()

	function a() {}

	function b() {}
	`
	_, err = compiler.New("", def)
	assert.Error(t, err)
}
