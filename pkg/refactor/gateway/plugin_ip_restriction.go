package gateway

import (
	"net"
	"net/http"
)

type ipRestrictionPlugin struct {
	conf ipRestrictionConfig
}

type ipRestrictionConfig struct {
	Whitelist []string `json:"whitelist" jsonschema:"required"`
}

func (ip ipRestrictionPlugin) build(c map[string]interface{}) (serve, error) {
	if err := unmarshalConfig(c, &ip.conf); err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "SplitHostPort failed", http.StatusInternalServerError)

			return false
		}
		if !isIPAllowed(clientIP, ip.conf.Whitelist) {
			http.Error(w, "Forbidden", http.StatusForbidden)

			return false
		}

		return true
	}, nil
}

func (ip ipRestrictionPlugin) getSchema() interface{} {
	return &ipRestrictionConfig{}
}

func isIPAllowed(clientIP string, allowedIPs []string) bool {
	for _, allowedIP := range allowedIPs {
		if clientIP == allowedIP {
			return true
		}
	}

	return false
}

//nolint:gochecknoinits
func init() {
	registry["ip_restriction_plugin"] = ipRestrictionPlugin{}
}
