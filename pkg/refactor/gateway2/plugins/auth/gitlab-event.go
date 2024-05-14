package auth

import (
	"context"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const (
	gitlabHeaderName = "X-Gitlab-Token"
)

type GitlabWebhookPlugin struct {
	Secret string `mapstructure:"secret"`
}

func (p *GitlabWebhookPlugin) NewInstance(_ core.EndpointV2, config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &GitlabWebhookPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *GitlabWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	// check request is already authenticated
	if gateway2.ParseRequestActiveConsumer(r) != nil {
		return r, nil
	}

	secret := r.Header.Get(gitlabHeaderName)
	if secret != p.Secret {
		return r, nil
	}

	c := &core.ConsumerFile{
		Username: "gitlab",
	}
	r = r.WithContext(context.WithValue(r.Context(), core.GatewayCtxKeyActiveConsumer, c))

	return r, nil
}

func (*GitlabWebhookPlugin) Type() string {
	return "gitlab-webhook-auth"
}

func init() {
	plugins.RegisterPlugin(&GitlabWebhookPlugin{})
}
