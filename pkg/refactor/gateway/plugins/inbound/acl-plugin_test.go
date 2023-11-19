package inbound_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestExecuteRequestACLGroupsPlugin(t *testing.T) {

	c := &core.Consumer{
		Username: "demo",
		Password: "hello",
		Tags:     []string{"tag1", "tag2"},
		Groups:   []string{"group1"},
	}

	p, _ := plugins.GetPluginFromRegistry(inbound.ACLPluginName)

	config := &inbound.ACLConfig{}
	p2, _ := p.Configure(config)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	p2.ExecutePlugin(context.Background(), c, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by fallback", w.Body.String())

	// test allow groups
	config = &inbound.ACLConfig{
		AllowGroups: []string{"group1"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config)
	p2.ExecutePlugin(context.Background(), c, w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	// test deny groups
	config = &inbound.ACLConfig{
		DenyGroups: []string{"group1"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config)
	p2.ExecutePlugin(context.Background(), c, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by group", w.Body.String())

	// test allow tags
	config = &inbound.ACLConfig{
		AllowTags: []string{"tag2"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config)
	p2.ExecutePlugin(context.Background(), c, w, r)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	// deny tag
	config = &inbound.ACLConfig{
		DenyTags: []string{"tag1"},
	}

	w = httptest.NewRecorder()
	p2, _ = p.Configure(config)
	p2.ExecutePlugin(context.Background(), c, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by tag", w.Body.String())

	// missing consumer
	w = httptest.NewRecorder()
	p2, _ = p.Configure(&inbound.ACLConfig{})
	p2.ExecutePlugin(context.Background(), nil, w, r)

	assert.Equal(t, http.StatusForbidden, w.Result().StatusCode)
	assert.Equal(t, "access denied by missing consumer", w.Body.String())
}
