package auth

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
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

func ConfigureGitlabWebhook(config interface{}, _ string) (core.PluginInstance, error) {
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

func (p *GitlabWebhookPlugin) ExecutePlugin(c *core.ConsumerFile, _ http.ResponseWriter, r *http.Request) bool {
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
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		GitlabWebhookPluginName,
		plugins.AuthPluginType,
		ConfigureGitlabWebhook))
}
