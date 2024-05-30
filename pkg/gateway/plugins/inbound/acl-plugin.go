package inbound

import (
	"context"
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway/plugins"
)

const (
	ACLPluginName = "acl"
)

// ACLConfig configures the ACL Plugin to allow, deny groups and tags.
type ACLConfig struct {
	AllowGroups []string `mapstructure:"allow_groups" yaml:"allow_groups"`
	DenyGroups  []string `mapstructure:"deny_groups"  yaml:"deny_groups"`
	AllowTags   []string `mapstructure:"allow_tags"   yaml:"allow_tags"`
	DenyTags    []string `mapstructure:"deny_tags"    yaml:"deny_tags"`
}

// ACLPlugin is a simple access control method. It checks the incoming consumer
// for tags and groups and allows or denies access.
type ACLPlugin struct {
	config *ACLConfig
}

func ConfigureACL(config interface{}, _ string) (core.PluginInstance, error) {
	aclConfig := &ACLConfig{}

	err := plugins.ConvertConfig(config, aclConfig)
	if err != nil {
		return nil, err
	}

	return &ACLPlugin{
		config: aclConfig,
	}, nil
}

func (acl *ACLPlugin) Config() interface{} {
	return acl.config
}

func (acl *ACLPlugin) Type() string {
	return ACLPluginName
}

func (acl *ACLPlugin) ExecutePlugin(c *core.ConsumerFile,
	w http.ResponseWriter, r *http.Request,
) bool {
	if c == nil {
		deny(r.Context(), "missing consumer", w)

		return false
	}

	if result(acl.config.AllowGroups, c.Groups) {
		return true
	}

	if result(acl.config.DenyGroups, c.Groups) {
		deny(r.Context(), "group", w)

		return false
	}

	if result(acl.config.AllowTags, c.Tags) {
		return true
	}

	if result(acl.config.DenyTags, c.Tags) {
		deny(r.Context(), "tag", w)

		return false
	}

	deny(r.Context(), "fallback", w)

	return false
}

func result(userValues []string, configValues []string) bool {
	for i := range userValues {
		g := userValues[i]
		for a := range configValues {
			if g == configValues[a] {
				return true
			}
		}
	}

	return false
}

func deny(ctx context.Context, t string, w http.ResponseWriter) {
	msg := fmt.Sprintf("access denied by %s", t)
	plugins.ReportError(ctx, w, http.StatusForbidden, msg, fmt.Errorf("forbidden"))
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		ACLPluginName,
		plugins.InboundPluginType,
		ConfigureACL))
}
