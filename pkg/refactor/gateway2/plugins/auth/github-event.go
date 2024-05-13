package auth

import (
	"bytes"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
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

func NewGithubWebhookPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	requestConvertConfig := &GithubWebhookPluginConfig{}

	err := gateway2.ConvertConfig(config, requestConvertConfig)
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

func (p *GithubWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	payload, err := github.ValidatePayload(r, []byte(p.config.Secret))
	if err != nil {
		slog.Error("can verify payload", "err", err)

		return r, nil
	}

	// reset body with payload
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	if c != nil {
		*c = core.ConsumerFile{
			Username: "github",
		}
	}

	return r, nil
}

func (*GithubWebhookPlugin) Type() string {
	return GithubWebhookPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(GithubWebhookPluginName, NewGithubWebhookPlugin)
}
