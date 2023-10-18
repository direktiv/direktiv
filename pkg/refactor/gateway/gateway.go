package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"

	"github.com/direktiv/direktiv/pkg/refactor/gateway/plugins"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type Handler struct {
	id           uuid.UUID
	objectPool   *sync.Map
	lock         sync.Mutex
	Status       string
	Error        error
	lastChecksum string
}

func NewHandler() *Handler {
	gw := &Handler{
		objectPool: &sync.Map{},
		lock:       sync.Mutex{},
		id:         uuid.New(),
	}

	return gw
}

func (gw *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("serving request", "route", r.Method+":"+r.URL.Path)
	p, _ := gw.objectPool.Load(r.Method + ":" + r.URL.Path)

	route, ok := p.(*plugins.Route)
	if !ok {
		// No plugin found, so we simply return.
		return
	}
	route.Execute(w, r)
}

func (gw *Handler) changeStatus(status string, err error) {
	gw.lock.Lock()
	defer gw.lock.Unlock()
	gw.Error = err
	gw.Status = status
}

func (gw *Handler) GetStatus() ConfigStatus {
	return ConfigStatus{
		ID:     gw.id.String(),
		Status: gw.Status,
		Error:  gw.Error,
	}
}

func (gw *Handler) replaceRoutes(routes []*plugins.Route) {
	check := generateChecksum(routes)
	if gw.lastChecksum == check {
		return
	}
	objectPool := &sync.Map{}
	for _, route := range routes {
		if err := route.Build(); err != nil {
			gw.changeStatus("Error building plugin instance", err)

			return
		}
		objectPool.Store(route.Method+":"+route.Path, route)
	}
	gw.lock.Lock()
	defer gw.lock.Unlock()
	gw.objectPool = objectPool
	gw.lastChecksum = check
}

func generateChecksum(routes []*plugins.Route) string {
	data := []byte(fmt.Sprintf("%v", routes))
	hash := sha256.New().Sum(data)

	return hex.EncodeToString(hash)
}

type ConfigStatus struct {
	ID string `json:"id"`
	plugins.RouteConfiguration
	Status string
	Error  error
}
