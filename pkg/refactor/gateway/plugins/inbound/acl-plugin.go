package inbound

import (
	"context"
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	ACLPluginName = "acl"
)

// ACLConfig configures the ACL Plugin to allow, deny groups and tags.
type ACLConfig struct {
	AllowGroups []string `yaml:"allos_groups"`
	DenyGroups  []string `yaml:"deny_groups"`
	AllowTags   []string `yaml:"allow_tags"`
	DenyTags    []string `yaml:"deny_tags"`
}

// ACLPlugin is a simple access control method. It checks the incoming consumer
// for tags and groups and allows or denies access.
type ACLPlugin struct {
	config *ACLConfig
}

func (acl ACLPlugin) Configure(config interface{}) (plugins.PluginInstance, error) {
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

func (acl ACLPlugin) Name() string {
	return ACLPluginName
}

func (acl ACLPlugin) Type() plugins.PluginType {
	return plugins.InboundPluginType
}

func (acl ACLPlugin) ExecutePlugin(ctx context.Context, c *core.Consumer,
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
	w.WriteHeader(http.StatusForbidden)

	// nolilnt
	w.Write([]byte(msg))
}

//nolint:gochecknoinits
func init() {
	plugins.AddPluginToRegistry(ACLPlugin{})
}
