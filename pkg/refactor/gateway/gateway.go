package gateway

import (
	"net/http"
	"sync"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type handler struct {
	lock *sync.Mutex

	endpoints []*core.Endpoint
}

func NewHandler() core.GatewayManager {
	return &handler{
		lock:      &sync.Mutex{},
		endpoints: make([]*core.Endpoint, 0),
	}
}

func (gw *handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	// nolint:errcheck
	_, _ = w.Write([]byte("hello gateway"))
}

func (gw *handler) SetEndpoints(list []*core.Endpoint) {
	gw.lock.Lock()
	defer gw.lock.Unlock()

	// clone list to the gateway status.
	newList := []*core.Endpoint{}
	for i := range list {
		cp := *list[i]
		newList = append(newList, &cp)
	}
	gw.endpoints = newList
}

func (gw *handler) ListEndpoints() []*core.Endpoint {
	newList := []*core.Endpoint{}
	for i := range gw.endpoints {
		cp := *gw.endpoints[i]
		newList = append(newList, &cp)
	}

	return newList
}

func (gw *handler) Start(done <-chan struct{}, wg *sync.WaitGroup) {
	go func() {
	loop:
		for {
			select {
			case <-done:
				break loop
			default:
			}
			gw.lock.Lock()
			gw.runCycle()
			gw.lock.Unlock()
			time.Sleep(time.Second)
		}

		wg.Done()
	}()
}

func (gw *handler) runCycle() {
	// TODO: construct plugins from list.
	time.Sleep(time.Millisecond)
}
