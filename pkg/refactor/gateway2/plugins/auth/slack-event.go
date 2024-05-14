package auth

import (
	"bytes"
	"encoding/json"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"io"
	"net/http"
	"net/url"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/slack-go/slack"
)

type SlackWebhookPlugin struct {
	Secret string `mapstructure:"secret"`
}

func (p *SlackWebhookPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &SlackWebhookPlugin{}

	err := gateway2.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *SlackWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	// check request is already authenticated
	if gateway2.ExtractContextActiveConsumer(r) != nil {
		return r
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		gateway2.WriteInternalError(r, w, err, "can not read request body")
		return nil
	}

	sv, err := slack.NewSecretsVerifier(r.Header, p.Secret)
	if err != nil {
		gateway2.WriteInternalError(r, w, err, "can not create slack verifier")
		return nil
	}

	if _, err := sv.Write(body); err != nil {
		gateway2.WriteInternalError(r, w, err, "can not write slack hmac")
		return nil
	}

	// hmac is not valid
	if err := sv.Ensure(); err != nil {
		gateway2.WriteInternalError(r, w, err, "slack hmac failed")
		return nil
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
			gateway2.WriteInternalError(r, w, err, "can parse url form encoded data")
			return nil
		}

		b, err := json.Marshal(v)
		if err != nil {
			gateway2.WriteInternalError(r, w, err, "can not marshal slack data")
			return nil
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
	} else {
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return r
}

func (*SlackWebhookPlugin) Type() string {
	return "slack-webhook-auth"
}

func init() {
	gateway2.RegisterPlugin(&SlackWebhookPlugin{})
}
