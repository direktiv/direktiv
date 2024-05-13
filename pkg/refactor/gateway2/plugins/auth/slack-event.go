package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func NewSlackWebhookPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
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

func (p *SlackWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	// check request is already authenticated
	if gateway2.ReadActiveConsumerFromContext(r) != nil {
		return r, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, fmt.Errorf("can not read request body")
	}

	sv, err := slack.NewSecretsVerifier(r.Header, p.config.Secret)
	if err != nil {
		slog.Error("can not create slack verifier", "err", err)
		return nil, fmt.Errorf("can not create slack verifier")
	}

	if _, err := sv.Write(body); err != nil {
		slog.Error("can not write slack hmac", "err", err)
		return nil, fmt.Errorf("can not write slack hmac")
	}

	// hmac is not valid
	if err := sv.Ensure(); err != nil {
		slog.Error("slack hmac failed", "err", err)
		return nil, fmt.Errorf("slack hmac failed")
	}

	c := &core.ConsumerFile{
		Username: "slack",
	}
	// set active comsumer.
	r = r.WithContext(context.WithValue(r.Context(), core.GatewayCtxKeyActiveConsumer, c))

	// convert to json if url encoded
	// nolint:canonicalheader
	if r.Header.Get("Content-type") == "application/x-www-form-urlencoded" {
		v, err := url.ParseQuery(string(body))
		if err != nil {
			slog.Error("can parse url form encoded data", "err", err)
			return nil, fmt.Errorf("can parse url form encoded data")
		}

		b, err := json.Marshal(v)
		if err != nil {
			slog.Error("can not marshal slack data", "err", err)
			return nil, fmt.Errorf("can not marshal slack data")
		}

		r.Body = io.NopCloser(bytes.NewBuffer(b))
	} else {
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return r, nil
}

func (*SlackWebhookPlugin) Type() string {
	return GithubWebhookPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(SlackWebhookPluginName, NewSlackWebhookPlugin)
}
