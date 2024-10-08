package gateway

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
	"github.com/pkg/errors"
)

// manager struct implements core.GatewayManager by wrapping a pointer to router struct. Whenever endpoint and
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

var _ core.GatewayManager = &manager{}

func NewManager(db *database.SQLStore) core.GatewayManager {
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
func (m *manager) SetEndpoints(list []core.Endpoint, cList []core.Consumer) error {
	cList = slices.Clone(cList)

	err := m.interpolateConsumersList(cList)
	if err != nil {
		return errors.Wrap(err, "interpolate consumer files")
	}
	newOne := buildRouter(list, cList)
	m.atomicSetRouter(newOne)

	return nil
}

// interpolateConsumersList translates matic consumer function "fetchSecret" in consumer files.
func (m *manager) interpolateConsumersList(list []core.Consumer) error {
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

func consumersForAPI(consumers []core.Consumer) any {
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
		if newItem.Errors == nil {
			newItem.Errors = []string{}
		}
		result = append(result, newItem)
	}

	return result
}

func endpointsForAPI(endpoints []core.Endpoint) any {
	type output struct {
		Methods        []string `json:"methods"`
		Path           string   `json:"path,omitempty"`
		AllowAnonymous bool     `json:"allow_anonymous"`
		PluginsConfig  struct {
			Auth     []core.PluginConfig `json:"auth,omitempty"`
			Inbound  []core.PluginConfig `json:"inbound,omitempty"`
			Target   core.PluginConfig   `json:"target,omitempty"`
			Outbound []core.PluginConfig `json:"outbound,omitempty"`
		} `json:"plugins"`
		Timeout    int      `json:"timeout"`
		FilePath   string   `json:"file_path"`
		Errors     []string `json:"errors"`
		ServerPath string   `json:"server_path"`
		Warnings   []string `json:"warnings"`
	}

	result := []any{}
	for _, item := range endpoints {
		newItem := output{
			Methods:        item.Methods,
			Path:           item.Path,
			AllowAnonymous: item.AllowAnonymous,
			Timeout:        item.Timeout,
			FilePath:       item.FilePath,
			Errors:         item.Errors,
		}
		newItem.PluginsConfig.Auth = item.PluginsConfig.Auth
		newItem.PluginsConfig.Inbound = item.PluginsConfig.Inbound
		newItem.PluginsConfig.Target = item.PluginsConfig.Target
		newItem.PluginsConfig.Outbound = item.PluginsConfig.Outbound

		newItem.Warnings = []string{}
		if newItem.Errors == nil {
			newItem.Errors = []string{}
		}
		// set server_path
		// TODO: remove this useless field
		if item.Path != "" {
			newItem.ServerPath = path.Clean(fmt.Sprintf("/ns/%s/%s", item.Namespace, item.Path))
		}
		result = append(result, newItem)
	}

	return result
}
