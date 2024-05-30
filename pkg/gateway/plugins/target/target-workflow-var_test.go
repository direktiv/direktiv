package target_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/target"
	"github.com/stretchr/testify/assert"
)

func TestConfigTargetFlowVarPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(target.TargetFlowVarPluginName)

	config := &target.WorkflowVarConfig{
		Flow:      "dummy",
		Variable:  "var",
		Namespace: "somerandom",
	}

	_, err := p.Configure(config, core.SystemNamespace)
	assert.NoError(t, err)

	_, err = p.Configure(config, "someother")
	assert.Error(t, err)

	_, err = p.Configure(config, "somerandom")
	assert.NoError(t, err)

	config = &target.WorkflowVarConfig{}
	_, err = p.Configure(config, "somerandom")
	assert.Error(t, err)
}
