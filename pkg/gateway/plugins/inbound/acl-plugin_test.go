package inbound_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestConfigACLPlugin(t *testing.T) {
	config := &inbound.ACLConfig{
		AllowGroups: []string{"group1"},
		DenyGroups:  []string{"group2"},
		AllowTags:   []string{"tag1"},
		DenyTags:    []string{"tag2"},
	}

	p, _ := plugins.GetPluginFromRegistry(inbound.ACLPluginName)
	p2, _ := p.Configure(config, core.SystemNamespace)

	configOut := p2.Config().(*inbound.ACLConfig)

	assert.Equal(t, config.AllowGroups, configOut.AllowGroups)
	assert.Equal(t, config.DenyGroups, configOut.DenyGroups)
	assert.Equal(t, config.AllowTags, configOut.AllowTags)
	assert.Equal(t, config.DenyTags, configOut.DenyTags)
}

func TestExecuteRequestACLGroupsPlugin(t *testing.T) {
	c := &core.ConsumerFile{
		Username: "demo",
		Password: "hello",
		Tags:     []string{"tag1", "tag2"},
		Groups:   []string{"group1"},
	}

	p, _ := plugins.GetPluginFromRegistry(inbound.ACLPluginName)

	config := &inbound.ACLConfig{}
	p2, _ := p.Configure(config, core.SystemNamespace)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	p2.ExecutePlugin(c, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by fallback: forbidden", w.Body.String())

	// test allow groups
	config = &inbound.ACLConfig{
		AllowGroups: []string{"group1"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config, core.SystemNamespace)
	p2.ExecutePlugin(c, w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	// test deny groups
	config = &inbound.ACLConfig{
		DenyGroups: []string{"group1"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config, core.SystemNamespace)
	p2.ExecutePlugin(c, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by group: forbidden", w.Body.String())

	// test allow tags
	config = &inbound.ACLConfig{
		AllowTags: []string{"tag2"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config, core.SystemNamespace)
	p2.ExecutePlugin(c, w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	// deny tag
	config = &inbound.ACLConfig{
		DenyTags: []string{"tag1"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config, core.SystemNamespace)
	p2.ExecutePlugin(c, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by tag: forbidden", w.Body.String())

	// missing consumer
	w = httptest.NewRecorder()
	p2, _ = p.Configure(&inbound.ACLConfig{}, core.SystemNamespace)
	p2.ExecutePlugin(nil, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by missing consumer: forbidden", w.Body.String())
}
