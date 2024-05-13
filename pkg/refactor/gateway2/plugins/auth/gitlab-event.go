package auth

import (
	"context"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const (
	gitlabWebhookPluginName = "gitlab-webhook-auth"
	gitlabHeaderName        = "X-Gitlab-Token"
)

type GitlabWebhookPluginConfig struct {
	Secret string `mapstructure:"secret" yaml:"secret"`
}

type GitlabWebhookPlugin struct {
	config *GitlabWebhookPluginConfig
}

func NewGitlabWebhookPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	gitlabConfig := &GitlabWebhookPluginConfig{}

	err := plugins.ConvertConfig(config, gitlabConfig)
	if err != nil {
		return nil, err
	}

	return &GitlabWebhookPlugin{
		config: gitlabConfig,
	}, nil
}

func (p *GitlabWebhookPlugin) Config() interface{} {
	return p.config
}

func (p *GitlabWebhookPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	// check request is already authenticated
	if gateway2.ParseRequestActiveConsumer(r) != nil {
		return r, nil
	}

	secret := r.Header.Get(gitlabHeaderName)

	if secret != p.config.Secret {
		return r, nil
	}

	c := &core.ConsumerFile{
		Username: "gitlab",
	}
	r = r.WithContext(context.WithValue(r.Context(), core.GatewayCtxKeyActiveConsumer, c))

	return r, nil
}

func (*GitlabWebhookPlugin) Type() string {
	return githubWebhookPluginName
}

func init() {
	plugins.RegisterPlugin(gitlabWebhookPluginName, NewGitlabWebhookPlugin)
}
