package auth

import (
	"net/http"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/gateway"
)

const (
	gitlabHeaderName = "X-Gitlab-Token"
)

type GitlabWebhookPlugin struct {
	Secret string `mapstructure:"secret"`
}

func (p *GitlabWebhookPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &GitlabWebhookPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *GitlabWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	// check request is already authenticated
	if gateway.ExtractContextActiveConsumer(r) != nil {
		return w, r
	}

	secret := r.Header.Get(gitlabHeaderName)
	if secret != p.Secret {
		return w, r
	}

	c := &core.Consumer{
		ConsumerFile: core.ConsumerFile{
			Username: "gitlab",
		},
	}
	r = gateway.InjectContextActiveConsumer(r, c)

	return w, r
}

func (*GitlabWebhookPlugin) Type() string {
	return "gitlab-webhook-auth"
}

func init() {
	gateway.RegisterPlugin(&GitlabWebhookPlugin{})
}
