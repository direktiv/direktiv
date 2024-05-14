package auth

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
	"github.com/google/go-github/v57/github"
)

type GithubWebhookPlugin struct {
	Secret string `mapstructure:"secret"`
}

func (p *GithubWebhookPlugin) NewInstance(_ core.EndpointV2, config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &GithubWebhookPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *GithubWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	// check request is already authenticated
	if gateway2.ParseRequestActiveConsumer(r) != nil {
		return r, nil
	}

	payload, err := github.ValidatePayload(r, []byte(p.Secret))
	if err != nil {
		slog.Error("cannot verify payload", "err", err)

		return r, nil
	}

	// reset body with payload
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	c := &core.ConsumerFile{
		Username: "github",
	}
	r = r.WithContext(context.WithValue(r.Context(), core.GatewayCtxKeyActiveConsumer, c))

	return r, nil
}

func (*GithubWebhookPlugin) Type() string {
	return "github-webhook-auth"
}

func init() {
	plugins.RegisterPlugin(&GithubWebhookPlugin{})
}
