package gateway

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type rateLimitPlugin struct {
	conf rateLimitConfig
	ips  sync.Map
}

type rateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute" jsonschema:"required"`
	ResetInterval     time.Duration `json:"reset_interval"      jsonschema:"required"`
}

func (rl *rateLimitPlugin) build(c map[string]interface{}) (serve, error) {
	if err := unmarshalConfig(c, &rl.conf); err != nil {
		return nil, err
	}
	// Setup a ticker to reset the counts
	go func() {
		ticker := time.NewTicker(rl.conf.ResetInterval)
		for range ticker.C {
			rl.resetCounts()
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) bool {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)

			return false
		}

		val, _ := rl.ips.LoadOrStore(ip, new(int32))
		requestCount, ok := val.(*int32)
		if !ok {
			http.Error(w, "requestCount is not an int", http.StatusInternalServerError)

			return false
		}
		if atomic.AddInt32(requestCount, 1) > int32(rl.conf.RequestsPerMinute) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)

			return false
		}

		return true
	}, nil
}

func (rl *rateLimitPlugin) resetCounts() {
	rl.ips.Range(func(key, value any) bool {
		rl.ips.Delete(key)

		return true
	})
}

func (rl *rateLimitPlugin) getSchema() interface{} {
	return &rateLimitConfig{}
}

//nolint:gochecknoinits
func init() {
	registry["rate_limit_plugin"] = &rateLimitPlugin{}
}
