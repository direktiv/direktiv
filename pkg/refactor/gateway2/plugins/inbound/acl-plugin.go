package inbound

import (
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

// ACLPlugin is a simple access control method. It checks the incoming consumer
// for tags and groups and allows or denies access.
type ACLPlugin struct {
	AllowGroups []string `mapstructure:"allow_groups"`
	DenyGroups  []string `mapstructure:"deny_groups"`
	AllowTags   []string `mapstructure:"allow_tags"`
	DenyTags    []string `mapstructure:"deny_tags"`
}

func (acl *ACLPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &ACLPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (acl *ACLPlugin) Type() string {
	return "acl"
}

func (acl *ACLPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	c := plugins.ExtractContextActiveConsumer(r)
	if c == nil {
		plugins.WriteInternalError(r, w, nil, "missing consumer")
		return nil
	}
	if result(acl.AllowGroups, c.Groups) {
		return r
	}
	if result(acl.DenyGroups, c.Groups) {
		plugins.WriteInternalError(r, w, nil, "denied user groups")
		return nil
	}
	if result(acl.AllowTags, c.Tags) {
		return r
	}
	if result(acl.DenyTags, c.Tags) {
		plugins.WriteInternalError(r, w, nil, "denied user tags")
		return nil
	}

	plugins.WriteInternalError(r, w, nil, "denied user")
	return nil
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

func init() {
	plugins.RegisterPlugin(&ACLPlugin{})
}
