package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/direktiv/direktiv/pkg/core"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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

	// // gateway info endpoint
	// if strings.HasSuffix(r.URL.Path, "/spec") {
	// 	ns := chi.URLParam(r, "namespace")

	// 	expand := false
	// 	if r.URL.Query().Get("expand") != "" {
	// 		expand = true
	// 	}
	// 	if ns != "" {
	// 		//nolint:contextcheck
	// 		WriteJSON(w, gatewayForAPI(filterNamespacedGateways(inner.gateways, ns), ns, m.db.FileStore(), inner.endpoints))
	// 		return
	// 	}
	// }

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
		Spec     map[string]interface{} `json:"spec"`
		FilePath string                 `json:"file_path"`
		Errors   []string               `json:"errors"`
	}

	apiDoc, _ := newOpenAPIDoc(ns, "/virual", "", fileStore)
	gw := output{
		FilePath: "virtual",
		Errors:   make([]string, 0),
		Spec:     *apiDoc.doc.GetSpecInfo().SpecJSON,
	}

	// we always take the first one, even if there are more
	if len(gateways) > 0 {
		g := gateways[0]

		// set file path
		gw.FilePath = g.FilePath

		var docData map[string]interface{}
		err := yaml.Unmarshal(g.Base, &docData)
		if err != nil {
			slog.Error("can not unmarshal gateway data", slog.Any("err", err),
				slog.String("namespace", ns))
			gw.Errors = append(gw.Errors, err.Error())
			return gw
		}

		endpointList := make(map[string]interface{})
		for i := range endpoints {
			_, errs := validateEndpoint(endpoints[i], ns, fileStore)
			if len(errs) > 0 {
				slog.Info("skipping endpoint with errors",
					slog.String("endpoint", endpoints[i].FilePath),
					slog.String("namespace", ns))
				continue
			}
			rel, err := filepath.Rel(filepath.Dir(g.FilePath), endpoints[i].FilePath)
			if err != nil {
				slog.Info("skipping endpoint with uncalculated path",
					slog.String("endpoint", endpoints[i].FilePath),
					slog.String("namespace", ns))
				continue
			}
			ref := make(map[string]string)
			ref["$ref"] = rel
			endpointList[endpoints[i].Config.Path] = ref
		}

		docData["paths"] = endpointList

		docBytes, err := yaml.Marshal(docData)
		if err != nil {
			slog.Error("can not marshal gateway data", slog.Any("err", err),
				slog.String("namespace", ns))
			gw.Errors = append(gw.Errors, err.Error())
			return gw
		}

		fmt.Println(string(docBytes))

		apiDoc, err = newOpenAPIDoc(ns, g.FilePath, string(docBytes), fileStore)
		if err != nil {
			slog.Error("gateway file invalid", slog.Any("err", err),
				slog.String("namespace", ns))
			gw.Errors = append(gw.Errors, err.Error())
			return gw
		}
		// gw.Spec = *apiDoc.doc.GetSpecInfo().SpecJSON

		// errs := apiDoc.validate()
		// gw.Errors = append(gw.Errors, errs...)

		if expand {
			spec, err := apiDoc.expand()
			if err != nil {
				slog.Error("gateway exapnd failed", slog.Any("err", err),
					slog.String("namespace", ns))
				gw.Errors = append(gw.Errors, err.Error())
				return gw
			}

			gw.Spec = spec
		}
	}

	// if there are more, it is an error
	if len(gateways) > 1 {
		f := make([]string, 0)
		for i := range gateways {
			f = append(f, gateways[i].FilePath)
		}
		gw.Errors = append(gw.Errors,
			fmt.Sprintf("multiple gateway specifications found: %s but using %s.", strings.Join(f, ", "), gw.FilePath))
	}

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

		pathItem, errors := validateEndpoint(item, ns, fileStore)
		newItem.Errors = append(newItem.Errors, errors...)
		newItem.Spec = pathItem

		if item.Config.Path != "" {
			newItem.ServerPath = path.Clean(fmt.Sprintf("/ns/%s/%s", item.Namespace, item.Config.Path))
		}
		result = append(result, newItem)
	}

	return result
}

func validateEndpoint(item core.Endpoint, ns string, fileStore filestore.FileStore) (interface{}, []string) {
	validationErrors := make([]string, 0)

	docString := fmt.Sprintf("openapi: 3.0.0\ninfo:\n   title: %s\n   version: \"1.0.0\"\npaths:\n   %s:\n      %s",
		ns, item.Config.Path, strings.ReplaceAll(string(item.Base), "\n", "\n      "))
	d, err := newOpenAPIDoc(ns, item.Config.Path, docString, fileStore)
	if err != nil {
		validationErrors = append(validationErrors, err.Error())
		return nil, validationErrors
	}

	errs := d.validate()
	validationErrors = append(validationErrors, errs...)

	mm := *d.doc.GetSpecInfo().SpecJSON

	paths, ok := mm["paths"]
	if !ok {
		validationErrors = append(validationErrors, "invalid pathitem layout")
		return nil, validationErrors
	}

	m, ok := paths.(map[string]interface{})
	if !ok {
		validationErrors = append(validationErrors, "invalid pathitem layout")
		return nil, validationErrors
	}

	value, ok := m[item.Config.Path]
	if !ok {
		validationErrors = append(validationErrors, "invalid pathitem layout")
		return nil, validationErrors
	}

	return value, validationErrors
}
