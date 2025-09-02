package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/gateway"
	"github.com/slack-go/slack"
)

type SlackWebhookPlugin struct {
	Secret string `mapstructure:"secret"`
}

func (p *SlackWebhookPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &SlackWebhookPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *SlackWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	// check request is already authenticated
	if gateway.ExtractContextActiveConsumer(r) != nil {
		return w, r
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not read request body")
		return nil, nil
	}

	sv, err := slack.NewSecretsVerifier(r.Header, p.Secret)
	if err != nil {
		gateway.WriteInternalError(r, w, err, "can not create slack verifier")
		return nil, nil
	}

	if _, err := sv.Write(body); err != nil {
		gateway.WriteInternalError(r, w, err, "can not write slack hmac")
		return nil, nil
	}

	// hmac is not valid
	if err := sv.Ensure(); err != nil {
		gateway.WriteInternalError(r, w, err, "slack hmac failed")
		return nil, nil
	}

	c := &core.Consumer{
		ConsumerFile: core.ConsumerFile{
			Username: "slack",
		},
	}
	// set active comsumer.
	r = gateway.InjectContextActiveConsumer(r, c)

	// convert to json if url encoded
	// nolint:canonicalheader
	if r.Header.Get("Content-type") == "application/x-www-form-urlencoded" {
		v, err := url.ParseQuery(string(body))
		if err != nil {
			gateway.WriteInternalError(r, w, err, "can parse url form encoded data")
			return nil, nil
		}

		b, err := json.Marshal(v)
		if err != nil {
			gateway.WriteInternalError(r, w, err, "can not marshal slack data")
			return nil, nil
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
	} else {
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return w, r
}

func (*SlackWebhookPlugin) Type() string {
	return "slack-webhook-auth"
}

func init() {
	gateway.RegisterPlugin(&SlackWebhookPlugin{})
}
