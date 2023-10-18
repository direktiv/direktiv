package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"golang.org/x/exp/slog"
)

type Handler struct {
	list         []*RouteConfiguration
	objectPool   *sync.Map
	lock         sync.Mutex
	lastChecksum string
}

func NewHandler() *Handler {
	slog.Info("init handler")
	gw := &Handler{
		objectPool: &sync.Map{},
		lock:       sync.Mutex{},
		list:       make([]*RouteConfiguration, 0),
	}

	return gw
}

func (gw *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prefix := "/api/v2/gateway"
	slog.Info("serving request", "route", r.Method+":"+r.URL.Path)
	if r.Method == "GET" && r.URL.Path == prefix {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		payLoad := struct {
			Data any `json:"data"`
		}{
			Data: gw.getEndpoints(),
		}
		err := json.NewEncoder(w).Encode(payLoad)
		if err != nil {
			slog.Error("error encoding", "route", r.Method+":"+r.URL.Path, "error", err)
		}

		return
	}
	path, _ := strings.CutPrefix(r.URL.Path, prefix)
	p, _ := gw.objectPool.Load(r.Method + ":" + path)

	route, ok := p.(*Route)
	if !ok {
		// No plugin found, so we simply return.
		return
	}
	route.execute(w, r)
}

func (gw *Handler) getEndpoints() []RouteConfiguration {
	gw.lock.Lock()
	defer gw.lock.Unlock()
	res := make([]RouteConfiguration, len(gw.list))
	for i, rc := range gw.list {
		res[i] = *rc
	}

	return res
}

func (gw *Handler) SetRoutes(list []*RouteConfiguration) {
	gw.lock.Lock()
	defer gw.lock.Unlock()
	gw.list = list
}

func (gw *Handler) replaceRoutes(routes []*RouteConfiguration) {
	check := generateChecksum(routes)
	if gw.lastChecksum == check {
		return
	}
	gw.lastChecksum = check
	objectPool := &sync.Map{}
	for _, route := range gw.list {
		r := Route{
			RouteConfiguration: *route,
		}

		if err := r.build(); err != nil {
			route.Err = err
			route.Status = "failed"

			return
		}
		objectPool.Store(route.Method+":"+route.Path, r)
	}
	gw.lock.Lock()
	defer gw.lock.Unlock()
	gw.objectPool = objectPool
}

func generateChecksum(routes []*RouteConfiguration) string {
	data := []byte(fmt.Sprintf("%v", routes))
	hash := sha256.New().Sum(data)

	return hex.EncodeToString(hash)
}

type ConfigStatus struct {
	*RouteConfiguration
}

func (gw *Handler) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	ticker := time.NewTicker(10 * time.Second) //nolint:gomnd
	defer ticker.Stop()

	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				wg.Done()

				return
			case <-ticker.C:
				gw.lock.Lock()
				gw.replaceRoutes(gw.list)
				gw.lock.Unlock()
			}
		}
	}()
}

type Route struct {
	RouteConfiguration
	plugins []plugins.Execute
}

type RouteConfiguration struct {
	Path           string
	Method         string
	Targets        plugins.Targets
	TimeoutSeconds int
	PluginsConfig  []plugins.Configuration
	Err            error
	Status         string
}

func (conf *Route) build() error {
	for _, c := range conf.PluginsConfig {
		exc, err := c.Build()
		if err != nil {
			return err
		}
		conf.plugins = append(conf.plugins, exc)
	}

	return nil
}

func (conf *Route) execute(w http.ResponseWriter, r *http.Request) {
	for _, p := range conf.plugins {
		if resp := p(w, r); resp.Status != http.StatusOK {
			return
		}
	}
}
