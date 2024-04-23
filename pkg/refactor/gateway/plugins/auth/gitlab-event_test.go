package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/auth"
	"github.com/stretchr/testify/assert"
)

const (
	correctGitlabPwd   = "secret"
	incorrectGitlabPwd = "incorrectsecret"
)

func TestGitlabEvent(t *testing.T) {
	c := &core.ConsumerFile{}

	config := auth.GitlabWebhookPluginConfig{
		Secret: correctGitlabPwd,
	}

	assert.True(t, executeGitlab(config, c, correctGitlabPwd))

	// consumer file is gitlab
	assert.Equal(t, c.Username, "gitlab")

	assert.False(t, executeGitlab(config, c, incorrectGitlabPwd))
}

func executeGitlab(config auth.GitlabWebhookPluginConfig, c *core.ConsumerFile, secret string) bool {
	p, _ := plugins.GetPluginFromRegistry(auth.GitlabWebhookPluginName)

	p2, _ := p.Configure(config, core.SystemNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", nil)
	r.Header.Add(auth.GitlabHeaderName, secret)

	w := httptest.NewRecorder()

	ret := p2.ExecutePlugin(c, w, r)

	return ret
}
