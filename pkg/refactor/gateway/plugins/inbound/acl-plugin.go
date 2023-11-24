package inbound

import (
	"context"
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	ACLPluginName = "acl"
)

// ACLConfig configures the ACL Plugin to allow, deny groups and tags.
type ACLConfig struct {
	AllowGroups []string `yaml:"allow_groups" mapstructure:"allow_groups"`
	DenyGroups  []string `yaml:"deny_groups" mapstructure:"deny_groups"`
	AllowTags   []string `yaml:"allow_tags" mapstructure:"allow_tags"`
	DenyTags    []string `yaml:"deny_tags" mapstructure:"deny_tags"`
}

// ACLPlugin is a simple access control method. It checks the incoming consumer
// for tags and groups and allows or denies access.
type ACLPlugin struct {
	config *ACLConfig
}

func ConfigureACL(config interface{}, ns string) (plugins.PluginInstance, error) {
	aclConfig := &ACLConfig{}

	if config != nil {
		err := mapstructure.Decode(config, &aclConfig)
		if err != nil {
			return nil, errors.Wrap(err, "configuration for target-flow invalid")
		}
	}

	return &ACLPlugin{
		config: aclConfig,
	}, nil
}

func (acl *ACLPlugin) Config() interface{} {
	return acl.config
}

func (acl *ACLPlugin) ExecutePlugin(ctx context.Context, c *spec.ConsumerFile,
	w http.ResponseWriter, r *http.Request) bool {

	if c == nil {
		deny("missing consumer", w)
		return false
	}

	if result(acl.config.AllowGroups, c.Groups) {
		return true
	}

	if result(acl.config.DenyGroups, c.Groups) {
		deny("group", w)
		return false
	}

	if result(acl.config.AllowTags, c.Tags) {
		return true
	}

	if result(acl.config.DenyTags, c.Tags) {
		deny("tag", w)
		return false
	}

	deny("fallback", w)
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

func deny(t string, w http.ResponseWriter) {
	msg := fmt.Sprintf("access denied by %s", t)
	plugins.ReportError(w, http.StatusForbidden, msg, fmt.Errorf("forbidden"))
}

func init() {
	plugins.AddPluginToRegistry(plugins.NewPluginBase(
		ACLPluginName,
		plugins.InboundPluginType,
		ConfigureACL))
}
