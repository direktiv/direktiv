package gateway2

import (
	"net/http"
	"sync/atomic"
	"unsafe"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type manager struct {
	routerPointer unsafe.Pointer
}

func (m *manager) atomicLoadRouter() *router {
	ptr := atomic.LoadPointer(&m.routerPointer)
	if ptr == nil {
		return nil
	}

	return (*router)(ptr)
}

func (m *manager) atomicSetRouter(inner *router) {
	atomic.StorePointer(&m.routerPointer, unsafe.Pointer(inner))
}

var _ core.GatewayManagerV2 = &manager{}

func NewManager() core.GatewayManagerV2 {
	return &manager{}
}

func (m *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inner := m.atomicLoadRouter()
	if inner == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "", "no active gateway endpoints")

		return
	}
	inner.serveMux.ServeHTTP(w, r)
}

func (m *manager) SetEndpoints(list []core.EndpointV2, cList []core.ConsumerV2) {
	newOne := buildRouter(list, cList)
	m.atomicSetRouter(newOne)
}

func (m *manager) ListEndpoints(namespace string) []core.EndpointV2 {
	inner := m.atomicLoadRouter()
	return filterNamespacedEndpoints(inner.endpoints, namespace)
}

func (m *manager) ListConsumers(namespace string) []core.ConsumerV2 {
	inner := m.atomicLoadRouter()
	return filterNamespacedConsumers(inner.consumers, namespace)
}
