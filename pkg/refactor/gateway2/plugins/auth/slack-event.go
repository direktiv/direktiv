package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
	"github.com/slack-go/slack"
)

type SlackWebhookPlugin struct {
	Secret string `mapstructure:"secret"`
}

func (p *SlackWebhookPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &SlackWebhookPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *SlackWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	// check request is already authenticated
	if gateway2.ExtractContextActiveConsumer(r) != nil {
		return r, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can not read request body", "err", err)
		return nil, fmt.Errorf("can not read request body")
	}

	sv, err := slack.NewSecretsVerifier(r.Header, p.Secret)
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

	c := &core.ConsumerV2{
		ConsumerFileV2: core.ConsumerFileV2{
			Username: "slack",
		},
	}
	// set active comsumer.
	r = gateway2.InjectContextActiveConsumer(r, c)

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
	return "slack-webhook-auth"
}

func init() {
	plugins.RegisterPlugin(&SlackWebhookPlugin{})
}
