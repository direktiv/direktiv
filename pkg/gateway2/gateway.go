package gateway2

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"slices"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/go-chi/chi/v5"
)

// manager struct implements core.GatewayManagerV2 by wrapping a pointer to router struct. Whenever endpoint and
// consumer files changes, method SetEndpoints should be called and this will build a new router and
// atomically swaps the old one.
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

// ServeHTTP makes this manager serves http requests.
func (m *manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inner := m.atomicLoadRouter()
	if inner == nil {
		WriteJSONError(w, http.StatusServiceUnavailable, "", "no active gateway endpoints")

		return
	}

	// setup /routes endpoint
	if strings.HasSuffix(r.URL.Path, "/routes") {
		ns := chi.URLParam(r, "namespace")
		if ns != "" {
			WriteJSON(w, endpointsForAPI(filterNamespacedEndpoints(inner.endpoints, ns, r.URL.Query().Get("path"))))
			return
		}
	}
	// setup /consumers endpoint
	if strings.HasSuffix(r.URL.Path, "/consumers") {
		ns := chi.URLParam(r, "namespace")
		if ns != "" {
			WriteJSON(w, consumersForAPI(filterNamespacedConsumers(inner.consumers, ns)))
			return
		}
	}

	inner.serveMux.ServeHTTP(w, r)
}

// SetEndpoints compiles a new router and atomically swaps the old one. No ongoing requests should be effected.
func (m *manager) SetEndpoints(list []core.EndpointV2, cList []core.ConsumerV2) {
	cList = slices.Clone(cList)

	err := m.interpolateConsumersList(cList)
	if err != nil {
		panic("TODO: unhandled error: " + err.Error())
	}
	newOne := buildRouter(list, cList)
	m.atomicSetRouter(newOne)
}

// interpolateConsumersList translates matic consumer function "fetchSecret" in consumer files.
func (m *manager) interpolateConsumersList(list []core.ConsumerV2) error {
	db, err := m.db.BeginTx(context.Background())
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer db.Rollback()

	for i, c := range list {
		c.Password, err = fetchSecret(db, c.Namespace, c.Password)
		if err != nil {
			c.Errors = append(c.Errors, fmt.Sprintf("couldn't fetch secret %s", c.Password))
			continue
		}

		c.APIKey, err = fetchSecret(db, c.Namespace, c.APIKey)
		if err != nil {
			c.Errors = append(c.Errors, fmt.Sprintf("couldn't fetch secret %s", c.APIKey))
			continue
		}
		list[i] = c
	}

	return nil
}

func consumersForAPI(consumers []core.ConsumerV2) any {
	type output struct {
		Username string   `json:"username"`
		Password string   `json:"password"`
		APIKey   string   `json:"api_key"`
		Tags     []string `json:"tags"`
		Groups   []string `json:"groups"`
		FilePath string   `json:"file_path"`
		Errors   []string `json:"errors"`
	}
	result := []any{}
	for _, item := range consumers {
		newItem := output{
			Username: item.Username,
			Password: item.Password,
			APIKey:   item.APIKey,
			Tags:     item.Tags,
			Groups:   item.Groups,
			FilePath: item.FilePath,
			Errors:   item.Errors,
		}
		result = append(result, newItem)
	}

	return result
}

func endpointsForAPI(endpoints []core.EndpointV2) any {
	type output struct {
		Methods        []string             `json:"methods"`
		Path           string               `json:"path"`
		AllowAnonymous bool                 `json:"allow_anonymous"`
		PluginsConfig  core.PluginsConfigV2 `json:"plugins"`
		Timeout        int                  `json:"timeout"`
		FilePath       string               `json:"file_path"`
		Errors         []string             `json:"errors"`
		ServerPath     string               `json:"server_path"`
		Warnings       []string             `json:"warnings"`
	}

	result := []any{}
	for _, item := range endpoints {
		newItem := output{
			Methods:        item.Methods,
			Path:           item.Path,
			AllowAnonymous: item.AllowAnonymous,
			PluginsConfig:  item.PluginsConfig,
			Timeout:        item.Timeout,
			FilePath:       item.FilePath,
			Errors:         item.Errors,
		}

		newItem.Warnings = []string{}
		// set server_path
		// TODO: remove this useless field
		if item.Path != "" {
			newItem.ServerPath = path.Clean(fmt.Sprintf("/ns/%s/%s", item.Namespace, item.Path))
		}

		result = append(result, newItem)
	}

	return result
}
