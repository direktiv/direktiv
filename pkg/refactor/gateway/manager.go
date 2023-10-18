package gateway

import (
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/direktiv/direktiv/pkg/refactor/spec"
)

type Manager struct {
	list    []*plugins.Route
	handler *Handler
	lock    sync.Mutex
}

func NewManager(handler *Handler) Manager {
	return Manager{
		handler: handler,
		lock:    sync.Mutex{},
	}
}

func (m *Manager) SetRoutes(list []*spec.PluginRouteFile) {
	m.lock.Lock()
	defer m.lock.Unlock()

	res := make([]*plugins.Route, 0, len(list))
	for _, e := range list {
		// Convert Targets
		var targets []plugins.Target
		for _, t := range e.Targets {
			targets = append(targets, plugins.Target{
				Method: t.Method,
				Host:   t.Host,
				Path:   t.Path,
				Scheme: t.Scheme,
			})
		}

		// Convert PluginsConfig
		var pluginsConfig []plugins.Configuration
		for _, pc := range e.PluginsConfig {
			pluginsConfig = append(pluginsConfig, plugins.Configuration{
				Name:                    pc.Name,
				Version:                 pc.Version,
				Comment:                 pc.Comment,
				Type:                    pc.Type,
				Priority:                pc.Priority,
				ExecutionTimeoutSeconds: pc.ExecutionTimeoutSeconds,
				RuntimeConfig:           pc.RuntimeConfig,
			})
		}

		routeConfig := plugins.RouteConfiguration{
			Path:           e.Path,
			Method:         e.Method,
			Targets:        targets,
			TimeoutSeconds: e.TimeoutSeconds,
			PluginsConfig:  pluginsConfig,
		}

		res = append(res, &plugins.Route{
			RouteConfiguration: routeConfig,
		})
	}
	m.list = res
}

func (m *Manager) Start(done <-chan struct{}, wg *sync.WaitGroup) {
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
				m.lock.Lock()
				m.runCycle()
				m.lock.Unlock()
			}
		}
	}()
}

func (m *Manager) runCycle() {
	m.handler.replaceRoutes(m.list)
}
