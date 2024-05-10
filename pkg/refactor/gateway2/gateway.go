package gateway2

import (
	"net/http"
	"sync/atomic"
	"unsafe"

	"github.com/direktiv/direktiv/pkg/refactor/core"
)

type manager struct {
	inner unsafe.Pointer
}

func (m *manager) loadInner() *immutableManager {
	ptr := atomic.LoadPointer(&m.inner)
	if ptr == nil {
		return nil
	}

	return (*immutableManager)(ptr)
}

func (m *manager) setInner(inner *immutableManager) {
	atomic.StorePointer(&m.inner, unsafe.Pointer(inner))
}

var _ core.GatewayManagerV2 = &manager{}

func NewManager() core.GatewayManagerV2 {
	return &manager{}
}

func (m *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inner := m.loadInner()
	if inner == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "", "no active gateway endpoints")

		return
	}
	inner.router.ServeHTTP(w, r)
}

func (m *manager) SetEndpoints(list []core.EndpointV2, cList []core.ConsumerV2) {
	newOne := newManager(list, cList)
	m.setInner(newOne)
}

func (m *manager) ListEndpoints(namespace string) []core.EndpointV2 {
	inner := m.loadInner()
	return filterNamespacedEndpoints(inner.endpoints, namespace)
}

func (m *manager) ListConsumers(namespace string) []core.ConsumerV2 {
	inner := m.loadInner()
	return filterNamespacedConsumers(inner.consumers, namespace)
}
