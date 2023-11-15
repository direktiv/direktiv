package gateway

import (
	"fmt"
	"net/http"
)

type sizeLimitPlugin struct {
	conf sizeLimitConfig
}

type sizeLimitConfig struct {
	MaxSize int64 `json:"max_size" jsonschema:"required"`
}

func (s sizeLimitPlugin) build(c map[string]interface{}) (serve, error) {
	if err := unmarshalConfig(c, &s.conf); err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) bool {
		if r.ContentLength > s.conf.MaxSize {
			http.Error(w, fmt.Sprintf("Request size exceeds the limit of %d bytes", s.conf.MaxSize), http.StatusRequestEntityTooLarge)

			return false
		}

		return true
	}, nil
}

func (s sizeLimitPlugin) getSchema() interface{} {
	return &sizeLimitConfig{}
}

//nolint:gochecknoinits
func init() {
	registry["size_limit_plugin"] = sizeLimitPlugin{}
}
