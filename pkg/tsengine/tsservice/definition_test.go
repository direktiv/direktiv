package tsservice_test

import (
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/stretchr/testify/assert"
)

func TestDefaultDefinition(t *testing.T) {

	emptyDef := ``
	c, _ := tsservice.New("", emptyDef)
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
	_, err := tsservice.New("", def)
	assert.Error(t, err)

	def = `
	var test = b()

	function a() {}

	function b() {}
	`
	_, err = tsservice.New("", def)
	assert.Error(t, err)
}

func TestBasicDefinition(t *testing.T) {

	emptyDef := `const flow: DirektivFlow = {
		scale: [
			{
				min: 1
			}
		]
	};`
	c, _ := tsservice.New("", emptyDef)
	info, _ := c.CompileFlow()

	assert.Equal(t, 1, info.Definition.Scale[0].Min)
}
