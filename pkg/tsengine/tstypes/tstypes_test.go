package tstypes_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/stretchr/testify/assert"
)

func TestBasicDefinition(t *testing.T) {

	emptyDef := `const flow: DirektivFlow = {
		scale: [
			{
				min: 1
			}
		]
	};
	function value() {
		return "jens"
	}`
	c, err := tsservice.NewTSServiceCompiler("", "", emptyDef)
	if err != nil {
		t.Error(err)
		return
	}
	info, err := c.CompileFlow()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, 1, info.Definition.Scale[0].Min)
}
