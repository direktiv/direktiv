package target_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/gateway/plugins/target"
	"github.com/stretchr/testify/assert"
)

func TestConfigTargetFlowPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(target.FlowPluginName)

	config := &target.WorkflowConfig{
		Flow:      "dummy",
		Namespace: "somerandom",
	}

	_, err := p.Configure(config, core.SystemNamespace)
	assert.NoError(t, err)

	_, err = p.Configure(config, "someother")
	assert.Error(t, err)

	_, err = p.Configure(config, "somerandom")
	assert.NoError(t, err)

	// no flow set, should fail
	config = &target.WorkflowConfig{}
	_, err = p.Configure(config, core.SystemNamespace)
	assert.Error(t, err)
}
