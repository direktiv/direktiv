package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/gateway/plugins/auth"

	"github.com/stretchr/testify/assert"
)

func TestConfigBasicAuthPlugin(t *testing.T) {
	config := auth.BasicAuthConfig{
		AddUsernameHeader: true,
		AddTagsHeader:     true,
		AddGroupsHeader:   true,
	}

	p, _ := plugins.GetPluginFromRegistry(auth.BasicAuthPluginName)
	p2, _ := p.Configure(config, core.SystemNamespace)

	configOut := p2.Config().(*auth.BasicAuthConfig)
	assert.Equal(t, config.AddGroupsHeader, configOut.AddGroupsHeader)
	assert.Equal(t, config.AddTagsHeader, configOut.AddTagsHeader)
	assert.Equal(t, config.AddUsernameHeader, configOut.AddUsernameHeader)
}

func TestExecuteBasicAuthPluginConfigure(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(auth.BasicAuthPluginName)

	// configure with nil
	_, err := p.Configure(nil, core.SystemNamespace)
	assert.NoError(t, err)

	// configure with nonsense
	_, err = p.Configure("random", core.SystemNamespace)
	assert.Error(t, err)

	config := &auth.BasicAuthConfig{}
	_, err = p.Configure(config, core.SystemNamespace)
	assert.NoError(t, err)
}

func TestExecuteBasicAuthPluginNoConsumer(t *testing.T) {
	w := httptest.NewRecorder()
	p, _ := plugins.GetPluginFromRegistry(auth.BasicAuthPluginName)

	config := &auth.BasicAuthConfig{
		AddUsernameHeader: true,
	}

	pi, _ := p.Configure(config, core.SystemNamespace)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", nil)

	c := &core.ConsumerFile{}

	pi.ExecutePlugin(c, w, r)

	// no consumer set, header is empty
	assert.Empty(t, r.Header.Get(plugins.ConsumerUserHeader))
}

func TestExecuteBasicAuthPlugin(t *testing.T) {
	userName := "demo"
	tags := []string{"tag1", "tag2"}
	groups := []string{"group1"}

	// test set header
	_, r := runBasicAuthRequest("demo", "hello", true, true, true)
	assert.Equal(t, userName, r.Header.Get(plugins.ConsumerUserHeader))
	assert.Equal(t, strings.Join(tags, ","), r.Header.Get(plugins.ConsumerTagsHeader))
	assert.Equal(t, strings.Join(groups, ","), r.Header.Get(plugins.ConsumerGroupsHeader))

	_, r = runBasicAuthRequest("demo", "hello", false, false, false)
	assert.Empty(t, r.Header.Get(plugins.ConsumerUserHeader))
	assert.Empty(t, r.Header.Get(plugins.ConsumerTagsHeader))
	assert.Empty(t, r.Header.Get(plugins.ConsumerGroupsHeader))

	// test invalid user
	_, r = runBasicAuthRequest("doesnotexist", "hello", true, true, true)
	assert.Empty(t, r.Header.Get(plugins.ConsumerUserHeader))

	// test invalid password
	_, r = runBasicAuthRequest("demoo", "wrongpassword", true, true, true)
	assert.Empty(t, r.Header.Get(plugins.ConsumerUserHeader))
}

func runBasicAuthRequest(user, pwd string, c1, c2, c3 bool) (*httptest.ResponseRecorder, *http.Request) {
	consumerList := consumer.NewConsumerList()

	// prepare consumer
	cl := []*core.ConsumerFile{
		{
			Username: "demo",
			Password: "hello",
			Tags:     []string{"tag1", "tag2"},
			Groups:   []string{"group1"},
		},
	}
	consumerList.SetConsumers(cl)

	p, _ := plugins.GetPluginFromRegistry(auth.BasicAuthPluginName)
	config := &auth.BasicAuthConfig{
		AddUsernameHeader: c1,
		AddTagsHeader:     c2,
		AddGroupsHeader:   c3,
	}
	p2, _ := p.Configure(config, core.SystemNamespace)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/dummy", nil)
	r = r.WithContext(context.WithValue(r.Context(), plugins.ConsumersParamCtxKey, consumerList))
	r.SetBasicAuth(user, pwd)

	c := &core.ConsumerFile{}

	p2.ExecutePlugin(c, w, r)

	return w, r
}
