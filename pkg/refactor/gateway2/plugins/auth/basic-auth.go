package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2"
	"github.com/direktiv/direktiv/pkg/refactor/gateway2/plugins"
)

type BasicAuthPlugin struct {
	AddUsernameHeader bool `mapstructure:"add_username_header"`
	AddTagsHeader     bool `mapstructure:"add_tags_header"`
	AddGroupsHeader   bool `mapstructure:"add_groups_header"`
}

var _ core.PluginV2 = &BasicAuthPlugin{}

func (ba *BasicAuthPlugin) NewInstance(config core.PluginConfigV2) (core.PluginV2, error) {
	pl := &BasicAuthPlugin{}

	err := plugins.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (ba *BasicAuthPlugin) Execute(w http.ResponseWriter, r *http.Request) *http.Request {
	// check request is already authenticated
	if plugins.ExtractContextActiveConsumer(r) != nil {
		return r
	}
	user, pwd, ok := r.BasicAuth()
	// no basic auth provided
	if !ok {
		return r
	}

	consumerList := plugins.ExtractContextConsumersList(r)
	if consumerList == nil {
		return r
	}
	consumer := gateway2.FindConsumerByUser(consumerList, user)
	// no consumer matching auth name
	if consumer == nil {
		return r
	}

	// comparing passwords
	userHash := sha256.Sum256([]byte(user))
	pwdHash := sha256.Sum256([]byte(pwd))
	userHashExpected := sha256.Sum256([]byte(consumer.Username))
	pwdHashExpected := sha256.Sum256([]byte(consumer.Password))

	usernameMatch := subtle.ConstantTimeCompare(userHash[:], userHashExpected[:]) == 1
	passwordMatch := subtle.ConstantTimeCompare(pwdHash[:], pwdHashExpected[:]) == 1

	if usernameMatch && passwordMatch {
		// set active comsumer.
		r = plugins.InjectContextActiveConsumer(r, consumer)
		// set headers if configured.
		if ba.AddUsernameHeader {
			r.Header.Set(gateway2.ConsumerUserHeader, consumer.Username)
		}

		if ba.AddTagsHeader && len(consumer.Tags) > 0 {
			r.Header.Set(gateway2.ConsumerTagsHeader, strings.Join(consumer.Tags, ","))
		}

		if ba.AddGroupsHeader && len(consumer.Groups) > 0 {
			r.Header.Set(gateway2.ConsumerGroupsHeader, strings.Join(consumer.Groups, ","))
		}
	}

	return r
}

func (ba *BasicAuthPlugin) Type() string {
	return "basic-auth"
}

func init() {
	plugins.RegisterPlugin(&BasicAuthPlugin{})
}
