package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

// manager struct implements core.GatewayManager by wrapping a pointer to router struct. Whenever endpoint and
// consumer files changes, method SetEndpoints should be called and this will build a new router and
// atomically swaps the old one.
type manager struct {
	routerPointer unsafe.Pointer

	db *database.SQLStore
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
			WriteJSON(w, endpointsForAPI(filterNamespacedEndpoints(inner.endpoints, ns, r.URL.Query().Get("path")), ns, m.db.FileStore()))
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

	// gateway info endpoint
	if strings.HasSuffix(r.URL.Path, "/info") {
		ns := chi.URLParam(r, "namespace")
		if ns != "" {
			expand := false
			if r.URL.Query().Get("expand") != "" {
				expand = true
			}
			//nolint:contextcheck
			WriteJSON(w, gatewayForAPI(filterNamespacedGateways(inner.gateways, ns), ns, m.db.FileStore(), inner.endpoints, expand))
			return
		}
	}

	inner.serveMux.ServeHTTP(w, r)
}

// SetEndpoints compiles a new router and atomically swaps the old one. No ongoing requests should be effected.
func (m *manager) SetEndpoints(list []core.Endpoint, cList []core.Consumer,
	glist []core.Gateway,
) error {
	cList = slices.Clone(cList)

	err := m.interpolateConsumersList(cList)
	if err != nil {
		return errors.Wrap(err, "interpolate consumer files")
	}
	newOne := buildRouter(list, cList, glist)
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

func gatewayForAPI(gateways []core.Gateway, ns string, fileStore filestore.FileStore, endpoints []core.Endpoint, expand bool) any {
	type output struct {
		Spec     any      `json:"spec"`
		FilePath string   `json:"file_path"`
		Errors   []string `json:"errors"`
		Warnings []string `json:"warnings"`
	}

	gw := output{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// edfault gateway
	g := core.Gateway{
		Base:     []byte(fmt.Sprintf("openapi: 3.0.0\ninfo:\n   title: %s\n   version: \"1.0\"", ns)),
		FilePath: "virtual",
	}

	// we always take the first one, even if there are more
	if len(gateways) > 0 {
		g = gateways[0]
	}

	// set file path
	gw.FilePath = g.FilePath

	// if there are more, it is an error
	if len(gateways) > 1 {
		f := make([]string, 0)
		for i := range gateways {
			f = append(f, gateways[i].FilePath)
		}
		gw.Warnings = append(gw.Warnings,
			fmt.Sprintf("multiple gateway specifications found: %s but using %s.", strings.Join(f, ", "), gw.FilePath))
	}

	doc, err := loadDoc(g.Base, g.FilePath, ns, fileStore)
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
		return gw
	}

	// add endpoints
	doc.Paths = openapi3.NewPaths()
	for i := range endpoints {
		ep := endpoints[i]
		_, err := validateEndpoint(ep, ns, fileStore)
		if err != nil {
			slog.Warn("skipping endpoint because of errors",
				slog.String("endpoint", ep.FilePath))
			gw.Warnings = append(gw.Warnings, fmt.Sprintf("skipping endpoint %s", ep.FilePath))
			continue
		}

		rel, err := filepath.Rel(filepath.Dir(gw.FilePath), ep.FilePath)
		if err != nil {
			slog.Warn("skipping endpoint because of of filepath calculation",
				slog.String("endpoint", ep.FilePath))
			gw.Warnings = append(gw.Warnings, fmt.Sprintf("skipping endpoint %s", ep.FilePath))
			continue
		}

		doc.Paths.Set(ep.Config.Path, &openapi3.PathItem{
			Ref: rel,
		})
	}

	if expand {
		c, err := doc.MarshalJSON()
		if err != nil {
			gw.Errors = append(gw.Errors, err.Error())
			return gw
		}
		doc, err = loadDoc(c, g.FilePath, ns, fileStore)
		if err != nil {
			gw.Errors = append(gw.Errors, err.Error())
			return gw
		}
		doc.InternalizeRefs(context.Background(), nil)
	}

	a, err := doc.MarshalYAML()
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
		return gw
	}
	gw.Spec = a

	return gw
}

func endpointsForAPI(endpoints []core.Endpoint, ns string, fileStore filestore.FileStore) any {
	type output struct {
		Spec       interface{} `json:"spec"`
		FilePath   string      `json:"file_path"`
		Errors     []string    `json:"errors"`
		ServerPath string      `json:"server_path"`
		Warnings   []string    `json:"warnings"`
	}

	result := []any{}

	for _, item := range endpoints {
		newItem := output{
			FilePath: item.FilePath,
			Errors:   item.Errors,
		}

		newItem.Warnings = []string{}
		if newItem.Errors == nil {
			newItem.Errors = []string{}
		}

		pathItem, err := validateEndpoint(item, ns, fileStore)
		if err != nil {
			slog.Error("endpoint invalid", slog.Any("err", err),
				slog.String("namespace", ns))
			newItem.Errors = append(newItem.Errors, err.Error())
		}
		newItem.Spec = pathItem

		if item.Config.Path != "" {
			newItem.ServerPath = path.Clean(fmt.Sprintf("/ns/%s/%s", item.Namespace, item.Config.Path))
		}
		result = append(result, newItem)
	}

	return result
}

func validateEndpoint(item core.Endpoint, ns string, fileStore filestore.FileStore) (*openapi3.PathItem, error) {
	// generate basic document for validation
	doc := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   ns,
			Version: "1.0",
		},
	}

	var pi openapi3.PathItem
	err := pi.UnmarshalJSON(item.Base)
	if err != nil {
		return &pi, err
	}

	doc.Paths = openapi3.NewPaths(
		openapi3.WithPath(item.Config.Path, &pi))

	// marshal doc to push it into the loader
	b, err := doc.MarshalJSON()
	if err != nil {
		return &pi, err
	}

	doc, err = loadDoc(b, item.FilePath, ns, fileStore)
	if err != nil {
		return &pi, err
	}

	err = doc.Validate(context.Background())
	return &pi, err
}

func loadDoc(data []byte, filePath, ns string, fileStore filestore.FileStore) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		file, err := fileStore.ForNamespace(ns).GetFile(context.Background(), url.String())
		if err != nil {
			return nil, err
		}
		return fileStore.ForFile(file).GetData(context.Background())
	}

	rel, err := filepath.Rel("/", filePath)
	if err != nil {
		return nil, err
	}

	return loader.LoadFromDataWithPath(data, &url.URL{Path: rel})
}
