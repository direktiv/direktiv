package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/consumer"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/auth"

	"github.com/stretchr/testify/assert"
)

func TestExecuteBasicAuthPluginConfigure(t *testing.T) {

	p, _ := plugins.GetPluginFromRegistry(auth.BasicAuthPluginName)

	// configure with nil
	_, err := p.Configure(nil)
	assert.NoError(t, err)

	// configure with nonsense
	_, err = p.Configure("random")
	assert.Error(t, err)

	config := &auth.BasicAuthConfig{}
	_, err = p.Configure(config)
	assert.NoError(t, err)

}

func TestExecuteBasicAuthPluginNoConsumer(t *testing.T) {

	w := httptest.NewRecorder()
	p, _ := plugins.GetPluginFromRegistry(auth.BasicAuthPluginName)

	config := &auth.BasicAuthConfig{
		AddUsernameHeader: true,
	}

	pi, _ := p.Configure(config)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", nil)

	c := &core.Consumer{}

	pi.ExecutePlugin(r.Context(), c, w, r)

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

	// prepare consumer
	cl := []*core.Consumer{
		{
			Username: "demo",
			Password: "hello",
			Tags:     []string{"tag1", "tag2"},
			Groups:   []string{"group1"},
		},
	}
	consumer.SetConsumer(cl)

	p, _ := plugins.GetPluginFromRegistry(auth.BasicAuthPluginName)
	config := &auth.BasicAuthConfig{
		AddUsernameHeader: c1,
		AddTagsHeader:     c2,
		AddGroupsHeader:   c3,
	}
	p2, _ := p.Configure(config)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/dummy", nil)
	r.SetBasicAuth(user, pwd)

	c := &core.Consumer{}

	p2.ExecutePlugin(r.Context(), c, w, r)

	return w, r

}
