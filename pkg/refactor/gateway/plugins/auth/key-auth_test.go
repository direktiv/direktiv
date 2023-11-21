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

func TestExecuteKeyAuthPluginConfigure(t *testing.T) {

	p, _ := plugins.GetPluginFromRegistry(auth.KeyAuthPluginName)

	// configure with nil
	_, err := p.Configure(nil)
	assert.NoError(t, err)

	// configure with nonsense
	_, err = p.Configure("random")
	assert.Error(t, err)

	// fails for missing name for the api key
	config := &auth.KeyAuthConfig{}
	_, err = p.Configure(config)
	assert.NoError(t, err)

	config.KeyName = "testme"
	_, err = p.Configure(config)
	assert.NoError(t, err)

}

func TestExecuteKeyAuthPluginNoConsumer(t *testing.T) {

	w := httptest.NewRecorder()
	p, _ := plugins.GetPluginFromRegistry(auth.KeyAuthPluginName)

	config := &auth.KeyAuthConfig{
		AddUsernameHeader: true,
	}

	p2, _ := p.Configure(config)

	r, _ := http.NewRequest(http.MethodPost, "/dummy", nil)

	c := &core.Consumer{}

	p2.ExecutePlugin(r.Context(), c, w, r)

	// no consumer set, header is empty
	assert.Empty(t, r.Header.Get(plugins.ConsumerUserHeader))

}

func TestExecuteKeyAuthPlugin(t *testing.T) {

	userName := "demo"
	tags := []string{"tag1", "tag2"}
	groups := []string{"group1"}
	key := "mykey"

	// test set header
	_, r := runKeyAuthRequest(key, true, true, true)
	assert.Equal(t, userName, r.Header.Get(plugins.ConsumerUserHeader))
	assert.Equal(t, strings.Join(tags, ","), r.Header.Get(plugins.ConsumerTagsHeader))
	assert.Equal(t, strings.Join(groups, ","), r.Header.Get(plugins.ConsumerGroupsHeader))

	_, r = runKeyAuthRequest(key, false, false, false)
	assert.Empty(t, r.Header.Get(plugins.ConsumerUserHeader))
	assert.Empty(t, r.Header.Get(plugins.ConsumerTagsHeader))
	assert.Empty(t, r.Header.Get(plugins.ConsumerGroupsHeader))

	// test invalid key
	_, r = runKeyAuthRequest("doesnotexist", true, true, true)
	assert.Empty(t, r.Header.Get(plugins.ConsumerUserHeader))

}

func runKeyAuthRequest(key string, c1, c2, c3 bool) (*httptest.ResponseRecorder, *http.Request) {

	consumerList := consumer.NewConsumerList()

	// prepare consumer
	cl := []*core.Consumer{
		{
			Username: "demo",
			APIKey:   "mykey",
			Tags:     []string{"tag1", "tag2"},
			Groups:   []string{"group1"},
		},
	}
	consumerList.SetConsumers(cl)

	p, _ := plugins.GetPluginFromRegistry(auth.KeyAuthPluginName)
	config := &auth.KeyAuthConfig{
		AddUsernameHeader: c1,
		AddTagsHeader:     c2,
		AddGroupsHeader:   c3,
		KeyName:           "testapikey",
	}
	p2, _ := p.Configure(config)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/dummy", nil)
	r.Header.Add("testapikey", key)

	c := &core.Consumer{}
	p2.ExecutePlugin(r.Context(), c, w, r)

	return w, r

}
