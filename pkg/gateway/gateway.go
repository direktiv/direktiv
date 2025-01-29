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
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/go-chi/chi/v5"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
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
		fmt.Println("DO GATEWAY!!!")
		g := gateways[0]

		// set file path
		gw.FilePath = g.FilePath

		apiDoc, err := newOpenAPIDoc(ns, g.FilePath, string(g.Base), fileStore)
		if err != nil {

		}

		gw.Spec = *apiDoc.doc.GetSpecInfo().SpecJSON

		errors := apiDoc.validate()
		if len(errors) > 0 {
			gw.Errors = append(gw.Errors, errors...)
		}

		// doc, err := libopenapi.NewDocumentWithConfiguration(g.Base,
		// 	openapiDocConfig(fileStore, ns, g.FilePath))
		// if err != nil {
		// 	gw.Errors = append(gw.Errors, err.Error())
		// 	return gw
		// }

		// add paths

		// hlval, errs := validator.NewValidator(doc)
		// if len(errs) > 0 {
		// 	for i := range errs {
		// 		gw.Errors = append(gw.Errors, errs[i].Error())
		// 	}
		// 	return gw
		// }

		// _, valErrs := hlval.ValidateDocument()
		// if len(errs) > 0 {
		// 	for i := range valErrs {
		// 		gw.Errors = append(gw.Errors, valErrs[i].Error())
		// 	}
		// 	return gw
		// }

		// var m map[string]interface{}
		// b, err := doc.Render()
		// yaml.Unmarshal(b, &m)
		// if err != nil {
		// 	gw.Errors = append(gw.Errors, err.Error())
		// 	return gw
		// }

		// gw.Spec = m
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
		// Spec       *v3high.PathItem `json:"spec"`
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
			// PathItem: item.PathItem,
		}

		newItem.Warnings = []string{}
		if newItem.Errors == nil {
			newItem.Errors = []string{}
		}

		pathItem, errors := validateEndpoint(item, ns, fileStore)
		newItem.Errors = append(newItem.Errors, errors...)

		var m map[string]interface{}
		b, _ := pathItem.Render()
		yaml.Unmarshal(b, &m)

		// j, _ := json.Marshal(m)
		// json.Unmarshal(j, &m)

		newItem.Spec = m
		// f, _ := json.MarshalIndent(pathItem, "", "   ")
		// fmt.Println(string(f))
		// gg, _ := pathItem.MarshalYAML()
		// out, _ := yaml.Marshal(gg)

		// fmt.Printf("YAML2 %+v\n", string(out))
		// newItem.Spec = gg
		// TODO: remove this useless field
		if item.Config.Path != "" {
			newItem.ServerPath = path.Clean(fmt.Sprintf("/ns/%s/%s", item.Namespace, item.Config.Path))
		}
		result = append(result, newItem)
	}

	return result
}

func validateEndpoint(item core.Endpoint, ns string, fileStore filestore.FileStore) (*v3high.PathItem, []string) {

	validationErrors := make([]string, 0)

	// var (
	// 	idxNode yaml.Node
	// 	n       v3low.PathItem
	// )

	// // we have to create the rolodex manually
	// idxConfig := &index.SpecIndexConfig{
	// 	BasePath:          filepath.Dir(item.FilePath),
	// 	AllowRemoteLookup: true,
	// 	AvoidBuildIndex:   true,
	// 	AllowFileLookup:   true,
	// }

	// rolodex := index.NewRolodex(idxConfig)
	// rolodex.AddLocalFS("/", &direktivOpenAPIFS{
	// 	fileStore: fileStore,
	// 	ns:        ns,
	// 	// files:     make(map[string]index.RolodexFile),
	// })

	// err := yaml.Unmarshal(item.Base, &idxNode)
	// if err != nil {
	// 	validationErrors = append(validationErrors, err.Error())
	// 	return nil, validationErrors
	// }

	// rolodex.SetRootNode(&idxNode)
	// err = rolodex.IndexTheRolodex()
	// if err != nil {
	// 	validationErrors = append(validationErrors, err.Error())
	// 	return nil, validationErrors
	// }

	// idxConfig.Rolodex = rolodex
	// err = low.BuildModel(idxNode.Content[0], &n)
	// if err != nil {
	// 	validationErrors = append(validationErrors, err.Error())
	// 	return nil, validationErrors
	// }

	// err = n.Build(context.Background(), nil, idxNode.Content[0], rolodex.GetRootIndex())
	// if err != nil {
	// 	validationErrors = append(validationErrors, err.Error())
	// 	return nil, validationErrors
	// }

	// pathItem := v3high.NewPathItem(&n)

	// // gg, _ := pi2.MarshalYAML()
	// // out, _ := yaml.Marshal(gg)
	// // fmt.Printf("YAML %+v\n", string(out))

	// doc, _ := libopenapi.NewDocumentWithConfiguration([]byte("openapi: 3.0.0\ninfo:\n   title: dummy\n   version: \"1.0.0\"\n   paths:"), &datamodel.DocumentConfiguration{
	// 	AllowFileReferences:   true,
	// 	AllowRemoteReferences: true,
	// 	BasePath:              filepath.Dir(item.FilePath),
	// 	AvoidIndexBuild:       true,
	// 	LocalFS: &direktivOpenAPIFS{
	// 		fileStore: fileStore,
	// 		ns:        ns,
	// 	},
	// })

	// v3Model, errs := doc.BuildV3Model()
	// if len(errs) > 0 {
	// 	for i := range errs {
	// 		validationErrors = append(validationErrors, errs[i].Error())
	// 	}
	// 	return nil, validationErrors
	// }

	// v3Model.Model.Paths.PathItems.Set(item.Config.Path, pathItem)
	// _, doc, _, errs = doc.RenderAndReload()
	// if len(errs) > 0 {
	// 	for i := range errs {
	// 		validationErrors = append(validationErrors, errs[i].Error())
	// 	}
	// 	return nil, validationErrors
	// }

	// hlval, errs := validator.NewValidator(doc)
	// if len(errs) > 0 {
	// 	for i := range errs {
	// 		fmt.Printf(">>>> %v\n", errs[i])
	// 		validationErrors = append(validationErrors, errs[i].Error())
	// 	}
	// 	return nil, validationErrors
	// }

	// _, valErrs := hlval.ValidateDocument()
	// if len(valErrs) > 0 {
	// 	for i := range valErrs {
	// 		fmt.Printf(">>>> %v\n", valErrs[i])
	// 		validationErrors = append(validationErrors, valErrs[i].Error())
	// 	}
	// 	return nil, validationErrors
	// }

	gg := fmt.Sprintf("openapi: 3.0.0\ninfo:\n   title: %s\n   version: \"1.0.0\"\npaths:\n   %s:\n      %s", ns, item.Config.Path, strings.ReplaceAll(string(item.Base), "\n", "\n      "))
	d, err := newOpenAPIDoc(ns, item.Config.Path, gg, fileStore)
	fmt.Println(d)
	fmt.Println(err)

	// fmt.Println(strings.ReplaceAll(string(item.Base), "\n", "\n      "))
	// value := "      " + strings.ReplaceAll(string(item.Base), "\n", "\n      ")
	// fmt.Println(value)
	// model, errs := d.doc.BuildV3Model()
	// fmt.Println(errs)

	a, _ := d.doc.Serialize()
	fmt.Println(string(a))
	// fmt.Printf("RENDER %v\n", string(j))

	ee := d.validate()
	fmt.Println(ee)

	m, errs := d.doc.BuildV3Model()
	fmt.Println(errs)

	pi, erhasr := m.Model.Paths.PathItems.Get(item.Config.Path)
	fmt.Println(erhasr)

	return pi, validationErrors
}
