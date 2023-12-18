package auth

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/google/go-github/v57/github"
)

const (
	GithubWebhookPluginName = "github-webhook-auth"
)

type GithubWebhookPluginConfig struct {
	Secret string `mapstructure:"secret" yaml:"secret"`
}

type GithubWebhookPlugin struct {
	config *GithubWebhookPluginConfig
}

func ConfigureGithubWebhook(config interface{}, _ string) (core.PluginInstance, error) {
	requestConvertConfig := &GithubWebhookPluginConfig{}

	err := plugins.ConvertConfig(config, requestConvertConfig)
	if err != nil {
		return nil, err
	}

	return &GithubWebhookPlugin{
		config: requestConvertConfig,
	}, nil
}

func (p *GithubWebhookPlugin) Config() interface{} {
	return p.config
}

func (p *GithubWebhookPlugin) ExecutePlugin(c *core.ConsumerFile, _ http.ResponseWriter, r *http.Request) bool {
	payload, err := github.ValidatePayload(r, []byte(p.config.Secret))
	if err != nil {
		slog.Error("can verify payload",
			slog.String("error", err.Error()))

		return true
	}

	// reset body with payload
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	if c != nil {
		*c = core.ConsumerFile{
			Username: "github",
		}
	}

	return true
}

func (*GithubWebhookPlugin) Type() string {
	return GithubWebhookPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		GithubWebhookPluginName,
		plugins.AuthPluginType,
		ConfigureGithubWebhook))
}
