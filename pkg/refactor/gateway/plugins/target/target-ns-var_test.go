package target_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/target"
	"github.com/stretchr/testify/assert"
)

func TestConfigNSVarPlugin(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(target.TargetNamespaceVarPluginName)

	config := &target.NamespaceVarConfig{
		Variable:  "dummy",
		Namespace: "somerandom",
	}

	_, err := p.Configure(config, core.MagicalGatewayNamespace)
	assert.NoError(t, err)

	_, err = p.Configure(config, "someother")
	assert.Error(t, err)

	_, err = p.Configure(config, "somerandom")
	assert.NoError(t, err)

	config = &target.NamespaceVarConfig{}
	_, err = p.Configure(config, "somerandom")
	assert.Error(t, err)
}
