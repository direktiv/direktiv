package inbound

import (
	"net/http"
	"slices"

	"github.com/direktiv/direktiv/internal/core"
	"github.com/direktiv/direktiv/internal/gateway"
)

// ACLPlugin is a simple access control method. It checks the incoming consumer
// for tags and groups and allows or denies access.
type ACLPlugin struct {
	AllowGroups []string `mapstructure:"allow_groups"`
	DenyGroups  []string `mapstructure:"deny_groups"`
	AllowTags   []string `mapstructure:"allow_tags"`
	DenyTags    []string `mapstructure:"deny_tags"`
}

func (acl *ACLPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &ACLPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (acl *ACLPlugin) Type() string {
	return "acl"
}

func (acl *ACLPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	c := gateway.ExtractContextActiveConsumer(r)
	if c == nil {
		gateway.WriteInternalError(r, w, nil, "missing consumer")
		return nil, nil
	}
	if result(acl.DenyGroups, c.Groups) {
		gateway.WriteForbiddenError(r, w, nil, "denied user groups")
		return nil, nil
	}
	if result(acl.DenyTags, c.Tags) {
		gateway.WriteForbiddenError(r, w, nil, "denied user tags")
		return nil, nil
	}
	if result(acl.AllowGroups, c.Groups) {
		return w, r
	}
	if result(acl.AllowTags, c.Tags) {
		return w, r
	}

	gateway.WriteForbiddenError(r, w, nil, "denied user")

	return nil, nil
}

func result(userValues []string, configValues []string) bool {
	for i := range userValues {
		g := userValues[i]
		if slices.Contains(configValues, g) {
			return true
		}
	}

	return false
}

func init() {
	gateway.RegisterPlugin(&ACLPlugin{})
}
