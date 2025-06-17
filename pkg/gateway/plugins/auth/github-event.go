package auth

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
	"github.com/google/go-github/v57/github"
)

type GithubWebhookPlugin struct {
	Secret string `mapstructure:"secret"`
}

func (p *GithubWebhookPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &GithubWebhookPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *GithubWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	// check request is already authenticated
	if gateway.ExtractContextActiveConsumer(r) != nil {
		return w, r
	}

	payload, err := github.ValidatePayload(r, []byte(p.Secret))
	if err != nil {
		slog.Error("cannot verify payload", "err", err)

		return w, r
	}

	// reset body with payload
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	c := &core.Consumer{
		ConsumerFile: core.ConsumerFile{
			Username: "github",
		},
	}
	r = gateway.InjectContextActiveConsumer(r, c)

	return w, r
}

func (*GithubWebhookPlugin) Type() string {
	return "github-webhook-auth"
}

func init() {
	gateway.RegisterPlugin(&GithubWebhookPlugin{})
}
