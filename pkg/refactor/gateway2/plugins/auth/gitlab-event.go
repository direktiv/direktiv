package auth

import (
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

const (
	GitlabWebhookPluginName = "gitlab-webhook-auth"
	GitlabHeaderName        = "X-Gitlab-Token"
)

type GitlabWebhookPluginConfig struct {
	Secret string `mapstructure:"secret" yaml:"secret"`
}

type GitlabWebhookPlugin struct {
	config *GitlabWebhookPluginConfig
}

func NewGitlabWebhookPlugin(config core.PluginConfigV2) (core.PluginV2, error) {
	gitlabConfig := &GitlabWebhookPluginConfig{}

	err := gateway2.ConvertConfig(config, gitlabConfig)
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
	secret := r.Header.Get(GitlabHeaderName)

	if secret != p.config.Secret {
		return false
	}

	*c = core.ConsumerFile{
		Username: "gitlab",
	}

	return true
}

func (*GitlabWebhookPlugin) Type() string {
	return GithubWebhookPluginName
}

//nolint:gochecknoinits
func init() {
	plugins.RegisterPlugin(GitlabWebhookPluginName, NewGitlabWebhookPlugin)
}
