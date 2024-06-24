package tsservice_test

import (
	"testing"

	_ "embed"

	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/stretchr/testify/assert"
)

func TestBasicDefinition(t *testing.T) {
	emptyDef := `const flow: DirektivFlow = {
		scale: [
			{
				min: 3
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
	info, err := c.Parse()
	if err != nil {
		t.Error(err)
		return
	}
	// Assertions
	assert.Equal(t, "default", info.Definition.Type)  // Default type
	assert.Equal(t, "always", info.Definition.Store)  // Default store
	assert.True(t, info.Definition.JSON)              // Default JSON setting
	assert.Equal(t, "PT15M", info.Definition.Timeout) // Default timeout

	// Scale assertions (assuming you only expect one scale entry)
	assert.Equal(t, 1, len(info.Definition.Scale))
	assert.Equal(t, 3, info.Definition.Scale[0].Min)
	assert.Equal(t, 1, info.Definition.Scale[0].Max)              // Default max
	assert.Equal(t, "instances", info.Definition.Scale[0].Metric) // Default metric
	assert.Equal(t, 100, info.Definition.Scale[0].Value)          // Default value
	assert.Equal(t, "value", info.Definition.State)
}
