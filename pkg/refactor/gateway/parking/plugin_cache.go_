package gateway

import (
	"net/http"
	"time"
)

type cachePlugin struct {
	conf cacheConfig
}

type cacheConfig struct {
	CacheDuration time.Duration `json:"cache_duration" jsonschema:"required"`
}

func (cp *cachePlugin) build(c map[string]interface{}) (serve, error) {
	var conf cacheConfig

	if err := unmarshalConfig(c, &conf); err != nil {
		return nil, err
	}

	cp.conf = conf

	return func(w http.ResponseWriter, r *http.Request) bool {
		// TODO: implement me
		return true
	}, nil
}

func (cp *cachePlugin) getSchema() interface{} {
	return &cacheConfig{}
}

//nolint:gochecknoinits
func init() {
	registry["cache_plugin"] = &cachePlugin{}
}
