package inbound_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins/inbound"
	"github.com/stretchr/testify/assert"
)

func TestGithubEvent(t *testing.T) {
	config := inbound.GithubWebhookPluginConfig{
		Secret: "It's a Secret to Everybody",
	}
	assert.True(t, execute(config, "Hello, World!"))
	config = inbound.GithubWebhookPluginConfig{
		Secret: "It's a Secret to Everybody",
	}
	assert.True(t, execute(config, "Hello, World!"))
}

func TestGithubEventValidation(t *testing.T) {
	assert.False(t, execute(inbound.GithubWebhookPluginConfig{Secret: "bad secret"}, "Hello, World!"))
	assert.False(t, execute(inbound.GithubWebhookPluginConfig{Secret: "It's a Secret to Everybody"}, "Hello, World!BadBody"))
	assert.True(t, execute(inbound.GithubWebhookPluginConfig{Secret: "It's a Secret to Everybody"}, "Hello, World!"))
}

func TestGithubPluginPreservesBody(t *testing.T) {
	p, _ := plugins.GetPluginFromRegistry(inbound.GithubWebhookPluginName)
	config := inbound.GithubWebhookPluginConfig{
		Secret: "It's a Secret to Everybody",
		// ListenForType: []string{"sometype", "issues"},
	}
	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", strings.NewReader("Hello, World!"))
	r.Header.Add("X-GitHub-Delivery", "72d3162e-cc78-11e3-81ab-4c9367dc0958")
	r.Header.Add("X-Hub-Signature-256", "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17")
	r.Header.Add("User-Agent", "GitHub-Hookshot/044aadd")
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Content-Length", "6615")
	r.Header.Add("X-GitHub-Event", "issues")
	r.Header.Add("X-GitHub-Hook-ID", "292430182")
	r.Header.Add("X-GitHub-Hook-Installation-Target-ID", "79929171")
	r.Header.Add("X-GitHub-Hook-Installation-Target-Type", "repository")

	w := httptest.NewRecorder()
	b := p2.ExecutePlugin(nil, w, r)
	assert.True(t, b, "Plugin did not executed as it should")
	reader, err := r.GetBody()
	assert.Nil(t, err, "could not read the body")
	body, err := io.ReadAll(reader)
	assert.Nil(t, err, "could not read the body")
	assert.Equal(t, string(body), "Hello, World!")
}

func execute(config inbound.GithubWebhookPluginConfig, body string) bool {
	p, _ := plugins.GetPluginFromRegistry(inbound.GithubWebhookPluginName)

	p2, _ := p.Configure(config, core.MagicalGatewayNamespace)

	r, _ := http.NewRequest(http.MethodGet, "/dummy", strings.NewReader(body))
	r.Header.Add("X-GitHub-Delivery", "72d3162e-cc78-11e3-81ab-4c9367dc0958")
	r.Header.Add("X-Hub-Signature-256", "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17")
	r.Header.Add("User-Agent", "GitHub-Hookshot/044aadd")
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Content-Length", "6615")
	r.Header.Add("X-GitHub-Event", "issues")
	r.Header.Add("X-GitHub-Hook-ID", "292430182")
	r.Header.Add("X-GitHub-Hook-Installation-Target-ID", "79929171")
	r.Header.Add("X-GitHub-Hook-Installation-Target-Type", "repository")

	w := httptest.NewRecorder()
	return p2.ExecutePlugin(nil, w, r)
}
