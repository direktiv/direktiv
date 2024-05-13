package inbound

import (
	"context"
	"fmt"
	"net/http"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

// ACLPlugin is a simple access control method. It checks the incoming consumer
// for tags and groups and allows or denies access.
type ACLPlugin struct {
	AllowGroups []string `mapstructure:"allow_groups" yaml:"allow_groups"`
	DenyGroups  []string `mapstructure:"deny_groups"  yaml:"deny_groups"`
	AllowTags   []string `mapstructure:"allow_tags"   yaml:"allow_tags"`
	DenyTags    []string `mapstructure:"deny_tags"    yaml:"deny_tags"`
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

func (acl *ACLPlugin) Execute(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
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

func init() {
	plugins.RegisterPlugin(&ACLPlugin{})
}
