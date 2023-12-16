package inbound

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
	GithubWebhookPluginName = "github-event"
)

type GithubWebhookPluginConfig struct {
	Secret string `mapstructure:"secret"          yaml:"secret"`
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

func (p *GithubWebhookPlugin) ExecutePlugin(_ *core.ConsumerFile, w http.ResponseWriter, r *http.Request) bool {

	payload, err := github.ValidatePayload(r, []byte(p.config.Secret))

	if err != nil {
		slog.Error("can verify payload",
			slog.String("error", err.Error()))
		plugins.ReportError(w, http.StatusForbidden,
			"signature", err)

		return false
	}

	// reset body with payload
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	return true
}

func (*GithubWebhookPlugin) Type() string {
	return RequestConvertPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		GithubWebhookPluginName,
		plugins.InboundPluginType,
		ConfigureGithubWebhook))
}
