package gateway

import (
	"context"
	"fmt"
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
	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi/datamodel"
	basehigh "github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
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

type openAPIResolver struct {
	path, ns  string
	fileStore filestore.FileStore
}

func (r *openAPIResolver) resolveDirektivPath(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
	path := url.String()

	fmt.Printf("RESOLVING %v\n", url)

	// if not absolute we need to calculate path
	if !filepath.IsAbs(url.String()) {
		p, err := filepath.Rel(filepath.Dir(path),
			filepath.Join(filepath.Dir(path), url.String()))
		if err != nil {
			return nil, err
		}
		path = p
	}

	fmt.Printf("RESOLVING2 %v\n", path)

	file, err := r.fileStore.ForNamespace(r.ns).GetFile(context.Background(), path)
	if err != nil {
		return nil, err
	}

	return r.fileStore.ForFile(file).GetData(context.Background())
	// return nil, err
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
			//nolint:contextcheck
			WriteJSON(w, gatewayForAPI(filterNamespacedGateways(inner.gateways, ns), ns, m.db.FileStore(), inner.endpoints))
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

func defaultGateway(ns string) (libopenapi.Document, error) {
	defaultSpecHigh := &v3high.Document{
		Version: "3.0.0",
		Info: &basehigh.Info{
			Title:   ns,
			Version: "1.0",
		},
	}

	// render the document to YAML.
	yamlSpec, err := defaultSpecHigh.Render()
	if err != nil {
		return nil, err
	}

	return libopenapi.NewDocument(yamlSpec)
}

func gatewayForAPI(gateways []core.Gateway, ns string, fileStore filestore.FileStore, endpoints []core.Endpoint) any {
	type output struct {
		Spec     map[string]interface{}
		FilePath string   `json:"file_path"`
		Errors   []string `json:"errors"`
	}

	gw := output{
		FilePath: "virtual",
		Errors:   make([]string, 0),
	}

	doc, err := defaultGateway(ns)
	if err != nil {
		gw.Errors = append(gw.Errors, err.Error())
		return gw
	}

	// set default spec
	gw.Spec = *doc.GetSpecInfo().SpecJSON

	// TODO: add paths

	// we always take the first one, even if there are more
	if len(gateways) > 0 {
		g := gateways[0]

		// set file path
		gw.FilePath = g.FilePath

		g.Base.SetConfiguration(&datamodel.DocumentConfiguration{
			LocalFS: &DirektivOpenAPIFS{
				fileStore: fileStore,
				ns:        ns,
			},
			AllowFileReferences:   true,
			AllowRemoteReferences: false,
		})

		gw.Spec = *g.Base.GetSpecInfo().SpecJSON

		// // check base errors
		// model, errs := g.Base.BuildV3Model()
		// if len(errs) > 0 {
		// 	for i := range errs {
		// 		gw.Errors = append(gw.Errors, errs[i].Error())
		// 	}
		// 	return gw
		// }

		// // reset servers and paths
		// model.Model.Paths.PathItems = &orderedmap.Map[string, *v3high.PathItem]{}
		// model.Model.Servers = []*v3high.Server{}

		// g2, _ := g.Base.Serialize()
		// fmt.Println(string(g2))

		// _, rendered, _, renderErrs := g.Base.RenderAndReload()
		// if len(renderErrs) > 0 {
		// 	for i := range renderErrs {
		// 		gw.Errors = append(gw.Errors, renderErrs[i].Error())
		// 	}
		// 	return gw
		// }

		// g2, _ = rendered.Serialize()
		// fmt.Println(string(g2))

		// g.Base = rendered
		// gw.Spec = *g.Base.GetSpecInfo().SpecJSON

		// high level validation
		hlValidator, valErrors := validator.NewValidator(g.Base)
		if len(valErrors) > 0 {
			for i := range valErrors {
				gw.Errors = append(gw.Errors, valErrors[i].Error())
			}
			return gw
		}

		_, errsDoc := hlValidator.ValidateDocument()
		if len(errsDoc) > 0 {
			for i := range errsDoc {
				if len(errsDoc[i].SchemaValidationErrors) > 0 {
					for a := range errsDoc[i].SchemaValidationErrors {
						gw.Errors = append(gw.Errors, errsDoc[i].SchemaValidationErrors[a].Reason)
					}
				} else {
					gw.Errors = append(gw.Errors, errsDoc[i].Error())
				}
			}
			return gw
		}

		// gw.Spec = do.Model

		// fmt.Println(do.Model.Paths)

		// c := orderedmap.Iterate(context.Background(), do.Model.Paths.PathItems)
		// for pair := range c {
		// 	fmt.Println(pair.Key())
		// 	fmt.Println(pair.Value())
		// }
		// gg, err := do.Model.MarshalYAML()

		// o, err := yaml.Marshal(gg)
		// fmt.Printf("%v %v\n", string(o), err)

		// bb, err := bundler.BundleDocument(&do.Model)
		// fmt.Printf("%v %v\n", err, string(bb))

		// model, err := g.Base.BuildV3Model()

		// 	gw.Errors = g.Errors
		// 	gw.FilePath = g.FilePath
		// gw.Spec = g.RenderedBase

		// 	if gw.Spec.Info == nil {
		// 		gw.Spec.Info = &openapi3.Info{}
		// 	}

		// 	resolver := openAPIResolver{
		// 		path:      g.FilePath,
		// 		fileStore: fileStore,
		// 		ns:        ns,
		// 	}

		// 	loader := openapi3.NewLoader()
		// 	loader.IsExternalRefsAllowed = true
		// 	loader.ReadFromURIFunc = resolver.resolveDirektivPath

		// 	// if there is an error we set the default spec to avoid panics in unmarshal
		// 	err := loader.ResolveRefsIn(&gw.Spec, nil)
		// 	if err != nil {
		// 		fmt.Println("1")
		// 		fmt.Println(err)
		// 		gw.Spec = defaultSpec
		// 		gw.Errors = append(gw.Errors, err.Error())
		// 	}

		// 	err = gw.Spec.Validate(context.Background())
		// 	if err != nil {
		// 		fmt.Println("2")
		// 		fmt.Println(err)
		// 		gw.Spec = defaultSpec
		// 		gw.Errors = append(gw.Errors, err.Error())
		// 	}

		// 	// add routes
		// 	gw.Spec.Paths = openapi3.NewPathsWithCapacity(0)
		// 	for _, item := range endpoints {
		// 		verr := validateEndpoint(item, ns, fileStore)
		// 		if len(verr) == 0 {
		// 			rel, err := filepath.Rel(filepath.Dir(g.FilePath), item.FilePath)
		// 			if err != nil {
		// 				rel = item.FilePath
		// 			}
		// 			gw.Spec.Paths.Set(item.Config.Path, &openapi3.PathItem{
		// 				Ref: rel,
		// 			})

		// 			// b, err := resolver.resolveDirektivPath(loader, &url.URL{
		// 			// 	Path: rel,
		// 			// })

		// 			// fmt.Printf("PATH %v >%v<\n", err, string(b))

		// 		} else {
		// 			gw.Errors = append(gw.Errors, fmt.Sprintf("route %v had errors", item.FilePath))
		// 		}
		// 	}

		// 	err = loader.ResolveRefsIn(&gw.Spec, nil)

		// 	fmt.Printf("FINAL %v %v\n", err, gw.Spec.Paths.Find("/testme"))

		// 	b, err := gw.Spec.MarshalJSON()
		// 	fmt.Printf("FINAL2 %v\n", string(b))
		// if err != nil {
		// 	gw.Spec = defaultSpec
		// 	gw.Errors = append(gw.Errors, err.Error())
		// }

		// gw.Spec.Paths = openapi3.NewPaths(paths...)
		// gw.Spec.Paths = openapi3.NewPaths(openapi3.WithPath("/jsjjs", &openapi3.PathItem{
		// 	Ref: "whatveer",
		// }))

		// fmt.Printf("%+v\n", gw.Spec.Paths.Len())
		// fmt.Printf("%+v\n", gw.Spec.Paths.Value("/testme"))
		// fmt.Printf("THIRD %+v\n", gw.Spec.Paths.Map()["/testme"])
		// b, err := gw.Spec.Paths.MarshalJSON()
		// // o, err := yaml.Marshal(b)
		// fmt.Printf("TOVALUE %v %v\n", string(b), err)
		// // err = loader.ResolveRefsIn(&gw.Spec, nil)

		// gw.Spec.InternalizeRefs(context.Background(), func(t *openapi3.T, cr openapi3.ComponentRef) string {
		// 	fmt.Printf("DEFREF %v\n", cr)
		// 	return ""
		// })

		// gw.Spec.InternalizeRefs(context.Background(), func(t *openapi3.T, cr openapi3.ComponentRef) string {
		// 	fmt.Printf("DEFREF %v\n", cr)
		// 	return ""
		// })

		// b, err = gw.Spec.MarshalJSON()
		// fmt.Printf("%v %v\n", err, string(b))
		// // fmt.Printf("MAPVALUE %+v\n", gw.Spec.Paths.Map()["/testme"])

		// dummyDeleteme(gw.Spec, gw.Spec.Paths.Map(), func(t *openapi3.T, cr openapi3.ComponentRef) string {
		// 	fmt.Printf("DEFREF %v\n", cr)
		// 	return ""
		// })
	}

	// if there are more, it is an error
	// if len(gateways) > 1 {
	// 	f := make([]string, 0)
	// 	for i := range gateways {
	// 		f = append(f, gateways[i].FilePath)
	// 	}

	// 	gw.Errors = append(gw.Errors,
	// 		fmt.Sprintf("multiple gateway specifications found: %s but using %s.", strings.Join(f, ", "), gw.FilePath))
	// }

	return gw
}

func endpointsForAPI(endpoints []core.Endpoint, ns string, fileStore filestore.FileStore) any {
	type output struct {
		PathItem   v3high.PathItem `json:"path_item"`
		FilePath   string          `json:"file_path"`
		Errors     []string        `json:"errors"`
		ServerPath string          `json:"server_path"`
		Warnings   []string        `json:"warnings"`
	}

	result := []any{}

	// l := openapi3.NewLoader()
	// l.IsExternalRefsAllowed = true

	for _, item := range endpoints {
		newItem := output{
			FilePath: item.FilePath,
			Errors:   item.Errors,
			// PathItem: item.RenderedPathItem,
		}

		newItem.Warnings = []string{}
		if newItem.Errors == nil {
			newItem.Errors = []string{}
		}

		if len(item.Errors) == 0 {

		}

		fmt.Println("DOC!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!1")
		fmt.Println(item.Doc)

		fmt.Printf("%+v\n", item)

		newItem.PathItem = item.PathItem

		// item.Doc.SetConfiguration(&datamodel.DocumentConfiguration{
		// 	LocalFS: &DirektivOpenAPIFS{
		// 		fileStore: fileStore,
		// 		ns:        ns,
		// 	},
		// 	AllowFileReferences:   true,
		// 	AllowRemoteReferences: false,
		// })

		// hlValidator, valErrors := validator.NewValidator(item.Doc)
		// if len(valErrors) > 0 {
		// 	for i := range valErrors {
		// 		newItem.Errors = append(newItem.Errors, valErrors[i].Error())
		// 	}
		// 	return newItem
		// }

		// _, errsDoc := hlValidator.ValidateDocument()
		// if len(errsDoc) > 0 {
		// 	for i := range errsDoc {
		// 		if len(errsDoc[i].SchemaValidationErrors) > 0 {
		// 			for a := range errsDoc[i].SchemaValidationErrors {
		// 				newItem.Errors = append(newItem.Errors, errsDoc[i].SchemaValidationErrors[a].Reason)
		// 			}
		// 		} else {
		// 			newItem.Errors = append(newItem.Errors, errsDoc[i].Error())
		// 		}
		// 	}
		// 	return newItem
		// }

		// newItem.Errors = append(newItem.Errors, validateEndpoint(item, ns, fileStore)...)

		// // create fake doc for validation
		// doc := &openapi3.T{
		// 	Paths:   openapi3.NewPaths(openapi3.WithPath(item.FilePath, &item.RenderedPathItem)),
		// 	OpenAPI: "3.0.0",
		// 	Info: &openapi3.Info{
		// 		Title:   "dummy",
		// 		Version: "1.0.0",
		// 	},
		// }

		// l.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		// 	path := url.String()

		// 	// if not absolute we need to calculate path
		// 	if !filepath.IsAbs(url.String()) {
		// 		p, err := filepath.Rel(filepath.Dir(item.FilePath),
		// 			filepath.Join(filepath.Dir(item.FilePath), url.String()))
		// 		if err != nil {
		// 			return nil, err
		// 		}
		// 		path = p
		// 	}

		// 	file, err := fileStore.ForNamespace(ns).GetFile(context.Background(), path)
		// 	if err != nil {
		// 		return nil, err
		// 	}

		// 	return fileStore.ForFile(file).GetData(context.Background())
		// }

		// err := l.ResolveRefsIn(doc, nil)
		// if err != nil {
		// 	newItem.Errors = append(newItem.Errors, err.Error())
		// }

		// // validate the whole thing
		// err = newItem.PathItem.Validate(context.Background())
		// if err != nil {
		// 	newItem.Errors = append(newItem.Errors, err.Error())
		// }

		// set server_path
		// TODO: remove this useless field
		if item.Config.Path != "" {
			newItem.ServerPath = path.Clean(fmt.Sprintf("/ns/%s/%s", item.Namespace, item.Config.Path))
		}
		result = append(result, newItem)
	}

	return result
}

func validateEndpoint(item core.Endpoint, ns string, fileStore filestore.FileStore) []string {

	validationErrors := make([]string, 0)

	// l := openapi3.NewLoader()
	// l.IsExternalRefsAllowed = true

	// // create fake doc for validation
	// doc := &openapi3.T{
	// 	Paths:   openapi3.NewPaths(openapi3.WithPath(item.FilePath, &item.RenderedPathItem)),
	// 	OpenAPI: "3.0.0",
	// 	Info: &openapi3.Info{
	// 		Title:   "dummy",
	// 		Version: "1.0.0",
	// 	},
	// }

	// l.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
	// 	path := url.String()

	// 	// if not absolute we need to calculate path
	// 	if !filepath.IsAbs(url.String()) {
	// 		p, err := filepath.Rel(filepath.Dir(item.FilePath),
	// 			filepath.Join(filepath.Dir(item.FilePath), url.String()))
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		path = p
	// 	}

	// 	file, err := fileStore.ForNamespace(ns).GetFile(context.Background(), path)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return fileStore.ForFile(file).GetData(context.Background())
	// }

	// err := l.ResolveRefsIn(doc, nil)
	// if err != nil {
	// 	validationErrors = append(validationErrors, err.Error())
	// }

	// // validate the whole thing
	// err = item.RenderedPathItem.Validate(context.Background())
	// if err != nil {
	// 	validationErrors = append(validationErrors, err.Error())
	// }

	return validationErrors
}
