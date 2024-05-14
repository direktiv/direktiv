package gateway2

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync/atomic"
	"unsafe"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
)

type manager struct {
	routerPointer unsafe.Pointer
	db            *database.SQLStore
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

func NewManager(db *database.SQLStore) core.GatewayManagerV2 {
	return &manager{
		db: db,
	}
}

func (m *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inner := m.atomicLoadRouter()
	if inner == nil {
		WriteJSONError(w, http.StatusServiceUnavailable, "", "no active gateway endpoints")

		return
	}
	inner.serveMux.ServeHTTP(w, r)
}

func (m *manager) SetEndpoints(list []core.EndpointV2, cList []core.ConsumerV2) {
	cList = slices.Clone(cList)

	err := m.interpolateConsumersList(cList)
	if err != nil {
		panic("TODO: unhandled error: " + err.Error())
	}
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

func (m *manager) interpolateConsumersList(list []core.ConsumerV2) error {
	db, err := m.db.BeginTx(context.Background())
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer db.Rollback()

	for i, c := range list {
		c.Password, err = fetchSecret(db, c.Namespace, c.Password)
		if err != nil {
			c.Errors = append(c.Errors, fmt.Errorf("couldn't fetch secret %s", c.Password))
			continue
		}

		c.APIKey, err = fetchSecret(db, c.Namespace, c.APIKey)
		if err != nil {
			c.Errors = append(c.Errors, fmt.Errorf("couldn't fetch secret %s", c.APIKey))
			continue
		}
		list[i] = c
	}

	return nil
}

func ParseRequestConsumersList(r *http.Request) []core.ConsumerV2 {
	res := r.Context().Value(core.GatewayCtxKeyConsumers)
	if res == nil {
		return nil
	}
	consumerList, ok := res.([]core.ConsumerV2)
	if !ok {
		return nil
	}

	return consumerList
}

func ParseRequestActiveConsumer(r *http.Request) *core.ConsumerV2 {
	res := r.Context().Value(core.GatewayCtxKeyActiveConsumer)
	if res == nil {
		return nil
	}
	consumer, ok := res.(*core.ConsumerV2)
	if !ok {
		return nil
	}

	return consumer
}
