package auth

import (
	"bytes"
	"encoding/json"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
	"github.com/slack-go/slack"
)

const (
	SlackWebhookPluginName = "slack-webhook-auth"
)

type SlackWebhookPluginConfig struct {
	Secret string `mapstructure:"secret" yaml:"secret"`
}

type SlackWebhookPlugin struct {
	config *SlackWebhookPluginConfig
}

func ConfigureSlackWebhook(config interface{}, _ string) (core.PluginInstance, error) {
	slackWebhookConfig := &SlackWebhookPluginConfig{}

	err := gateway2.ConvertConfig(config, slackWebhookConfig)
	if err != nil {
		return nil, err
	}

	return &SlackWebhookPlugin{
		config: slackWebhookConfig,
	}, nil
}

func (p *SlackWebhookPlugin) Config() interface{} {
	return p.config
}

func (p *SlackWebhookPlugin) ExecutePlugin(c *core.ConsumerFile, _ http.ResponseWriter, r *http.Request) bool {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can not read slack body", "err", err)
		return false
	}

	sv, err := slack.NewSecretsVerifier(r.Header, p.config.Secret)
	if err != nil {
		slog.Error("can not create slack verifier", "err", err)
		return false
	}

	if _, err := sv.Write(body); err != nil {
		slog.Error("can not write slack hmac", "err", err)
		return false
	}

	// hmac is not valid
	if err := sv.Ensure(); err != nil {
		slog.Error("slack hmac failed", "err", err)
		return false
	}

	*c = core.ConsumerFile{
		Username: "slack",
	}

	// convert to json if url encoded
	// nolint:canonicalheader
	if r.Header.Get("Content-type") == "application/x-www-form-urlencoded" {
		v, err := url.ParseQuery(string(body))
		if err != nil {
			slog.Error("can parse url form encoded data", "err", err)
			return false
		}

		b, err := json.Marshal(v)
		if err != nil {
			slog.Error("can not marshal slack data", "err", err)
			return false
		}

		r.Body = io.NopCloser(bytes.NewBuffer(b))
	} else {
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return true
}

func (*SlackWebhookPlugin) Type() string {
	return GithubWebhookPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(SlackWebhookPluginName)
}
