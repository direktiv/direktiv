package gateway

import (
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
)

type Manager struct {
	list    []*plugins.Route
	handler *Handler
	lock    sync.Mutex
}

func NewManager(handler *Handler) *Manager {
	return &Manager{
		handler: handler,
		lock:    sync.Mutex{},
	}
}

func (m *Manager) SetRoutes(list []*plugins.Route) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.list = list
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
