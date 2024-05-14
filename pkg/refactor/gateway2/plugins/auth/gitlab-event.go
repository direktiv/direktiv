package auth

import (
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

func (p *GitlabWebhookPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &GitlabWebhookPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (p *GitlabWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	// check request is already authenticated
	if gateway2.ExtractContextActiveConsumer(r) != nil {
		return r
	}

	secret := r.Header.Get(gitlabHeaderName)
	if secret != p.Secret {
		return r
	}

	c := &core.ConsumerV2{
		ConsumerFileV2: core.ConsumerFileV2{
			Username: "gitlab",
		},
	}
	r = gateway2.InjectContextActiveConsumer(r, c)

	return r
}

func (*GitlabWebhookPlugin) Type() string {
	return "gitlab-webhook-auth"
}

func init() {
	plugins.RegisterPlugin(&GitlabWebhookPlugin{})
}
