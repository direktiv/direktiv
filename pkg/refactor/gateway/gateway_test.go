package gateway_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/gateway"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createNS(db *database.SQLStore, ns string) {
	ctx := context.Background()

	db.DataStore().Namespaces().Create(ctx, &datastore.Namespace{
		Name: ns,
	})
	_, _ = db.FileStore().CreateRoot(ctx, uuid.New(), ns)
}

var wf1 = `direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: instant-response
    configuration:
       status_code: 202
       status_message: "TEST"
methods: 
  - GET
  - POST
path: /test`

var wfAuth = `direktiv_api: endpoint/v1
plugins:
  auth:
  - type: "key-auth"
    configuration:
       key_name: secret
  target:
    type: instant-response
    configuration:
       status_code: 202
       status_message: "TEST"
methods: 
  - GET
path: /test`

var wfOutbound = `direktiv_api: endpoint/v1
allow_anonymous: true
plugins:
  target:
    type: instant-response
    configuration:
       status_code: 202
       status_message: content
  outbound:
    - type: js-outbound
      configuration:
         script: |
            log(input)
            input["Headers"].Add("demo", "value")
            input["Headers"].Add("demo2", "value2")
            input["Body"] = "changed"
            input["Code"] = 202
    - type: js-outbound
      configuration:
         script: |
            log(input)
            input["Headers"].Add("demo3", "value3")
methods: 
  - GET
path: /test`

var consumerAuth = `direktiv_api: "consumer/v1"
username: user
password: pwd
api_key: key
tags:
- tag1
groups:
- group1`

var timeout = `direktiv_api: endpoint/v1
allow_anonymous: true
timeout: 1
plugins:
  outbound:
    - type: js-outbound
      configuration:
       script: |
          sleep(2)
    - type: js-outbound
      configuration:
        script: |
          log(input)
          input["Headers"].Add("demo3", "value3")
methods: 
  - GET
  - POST
path: /test`

func TestBasicGateway(t *testing.T) {
	ns1 := "ns1"

	dbMock, _ := database.NewMockGorm()

	db := database.NewSQLStore(dbMock, "dummy")

	createNS(db, ns1)
	createNS(db, core.SystemNamespace)

	// create endpoint in magical and custom namespace
	db.FileStore().ForNamespace(ns1).CreateFile(context.Background(), "/test.yaml",
		filestore.FileTypeEndpoint, "application/yaml", []byte(wf1))

	db.FileStore().ForNamespace(core.SystemNamespace).CreateFile(context.Background(),
		"/test.yaml", filestore.FileTypeEndpoint, "application/yaml", []byte(wf1))

	gm := gateway.NewGatewayManager(db)
	gm.UpdateAll()

	// test namespace URL
	resp := doRequest(t, fmt.Sprintf("/ns/%s/test", ns1), nil, gm)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	// test special namespace URL
	resp = doRequest(t, "/gw/test", nil, gm)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	// deleting namespace should be 404
	gm.DeleteNamespace(ns1)
	resp = doRequest(t, fmt.Sprintf("/ns/%s/test", ns1), nil, gm)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAuthGateway(t *testing.T) {
	dbMock, _ := database.NewMockGorm()

	db := database.NewSQLStore(dbMock, "dummy")
	createNS(db, core.SystemNamespace)
	db.FileStore().ForNamespace(core.SystemNamespace).CreateFile(context.Background(),
		"/test.yaml", filestore.FileTypeEndpoint, "application/yaml", []byte(wfAuth))

	db.FileStore().ForNamespace(core.SystemNamespace).CreateFile(context.Background(),
		"/consumer.yaml", filestore.FileTypeConsumer, "application/yaml", []byte(consumerAuth))

	gm := gateway.NewGatewayManager(db)
	gm.UpdateAll()

	resp := doRequest(t, "/gw/test", nil, gm)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// set api key header
	h := make(http.Header)
	h.Set("secret", "key")
	resp = doRequest(t, "/gw/test", h, gm)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestOutputPlugins(t *testing.T) {
	dbMock, _ := database.NewMockGorm()

	db := database.NewSQLStore(dbMock, "dummy")
	createNS(db, core.SystemNamespace)
	db.FileStore().ForNamespace(core.SystemNamespace).CreateFile(context.Background(),
		"/test.yaml", filestore.FileTypeEndpoint, "application/yaml", []byte(wfOutbound))

	gm := gateway.NewGatewayManager(db)
	gm.UpdateAll()

	resp := doRequest(t, "/gw/test", make(http.Header), gm)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "value", resp.Header.Get("demo"))
	assert.Equal(t, "value3", resp.Header.Get("demo3"))
}

func doRequest(t *testing.T, url string, headers http.Header, gm core.GatewayManager) *http.Response {
	router := chi.NewRouter()
	router.Handle("/gw/*", gm)
	router.Handle("/ns/{namespace}/*", gm)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", url, nil)

	rctx := chi.NewRouteContext()
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	r.Header = headers

	router.ServeHTTP(w, r)
	b, _ := httputil.DumpResponse(w.Result(), true)
	t.Log(string(b))

	return w.Result()
}

func TestTimeoutRequest(t *testing.T) {
	dbMock, _ := database.NewMockGorm()

	db := database.NewSQLStore(dbMock, "dummy")
	createNS(db, core.SystemNamespace)
	db.FileStore().ForNamespace(core.SystemNamespace).CreateFile(context.Background(),
		"/test.yaml", filestore.FileTypeEndpoint, "application/yaml", []byte(timeout))

	gm := gateway.NewGatewayManager(db)
	gm.UpdateAll()

	resp := doRequest(t, "/gw/test", make(http.Header), gm)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
}

func TestGetAllEndpoints(t *testing.T) {
	dbMock, _ := database.NewMockGorm()

	db := database.NewSQLStore(dbMock, "dummy")
	createNS(db, core.SystemNamespace)
	db.FileStore().ForNamespace(core.SystemNamespace).CreateFile(context.Background(),
		"/test.yaml", filestore.FileTypeEndpoint, "application/yaml", []byte(timeout))

	gm := gateway.NewGatewayManager(db)
	gm.UpdateAll()

	items, _ := gm.GetRoutes(core.SystemNamespace, "")
	assert.Equal(t, "/test", items[0].Path)
}
