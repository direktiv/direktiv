package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/gateway"
)

type BasicAuthPlugin struct {
	AddUsernameHeader bool `mapstructure:"add_username_header"`
	AddTagsHeader     bool `mapstructure:"add_tags_header"`
	AddGroupsHeader   bool `mapstructure:"add_groups_header"`
}

var _ core.Plugin = &BasicAuthPlugin{}

func (ba *BasicAuthPlugin) NewInstance(config core.PluginConfig) (core.Plugin, error) {
	pl := &BasicAuthPlugin{}

	err := gateway.ConvertConfig(config.Config, pl)
	if err != nil {
		return nil, err
	}

	return pl, nil
}

func (ba *BasicAuthPlugin) Execute(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
	// check request is already authenticated
	if gateway.ExtractContextActiveConsumer(r) != nil {
		return w, r
	}
	user, pwd, ok := r.BasicAuth()
	// no basic auth provided
	if !ok {
		return w, r
	}

	consumerList := gateway.ExtractContextConsumersList(r)
	if consumerList == nil {
		return w, r
	}
	consumer := gateway.FindConsumerByUser(consumerList, user)
	// no consumer matching auth name
	if consumer == nil {
		return w, r
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
		r = gateway.InjectContextActiveConsumer(r, consumer)
		// set headers if configured.
		if ba.AddUsernameHeader {
			r.Header.Set(gateway.ConsumerUserHeader, consumer.Username)
		}

		if ba.AddTagsHeader && len(consumer.Tags) > 0 {
			r.Header.Set(gateway.ConsumerTagsHeader, strings.Join(consumer.Tags, ","))
		}

		if ba.AddGroupsHeader && len(consumer.Groups) > 0 {
			r.Header.Set(gateway.ConsumerGroupsHeader, strings.Join(consumer.Groups, ","))
		}
	}

	return w, r
}

func (ba *BasicAuthPlugin) Type() string {
	return "basic-auth"
}

func init() {
	gateway.RegisterPlugin(&BasicAuthPlugin{})
}
