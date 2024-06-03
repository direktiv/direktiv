package tsengine_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/direktiv/direktiv/pkg/compiler"
	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/datastore"
	"github.com/direktiv/direktiv/pkg/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/tsengine"
	"github.com/direktiv/direktiv/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type testFile struct {
	path, mimetytpe, data string
	typ                   filestore.FileType
}

func testNamespace(db *gorm.DB, name string, secrets map[string]string, files []testFile) error {
	os.Setenv("DIREKTIV_JSENGINE_NAMESPACE", name)

	ds := datastoresql.NewSQLStore(db, "some_secret_key_")
	fs := filestoresql.NewSQLFileStore(db)

	ns, err := ds.Namespaces().Create(context.Background(), &datastore.Namespace{Name: name})
	if err != nil {
		return err
	}
	root, err := fs.CreateRoot(context.Background(), uuid.New(), name)
	if err != nil {
		return err
	}

	for k, v := range secrets {
		err = ds.Secrets().Set(context.Background(), &datastore.Secret{
			Name:      k,
			Namespace: ns.Name,
			Data:      []byte(v),
		})
		if err != nil {
			return err
		}
	}

	for i := range files {
		f := files[i]
		_, err := fs.ForRootID(root.ID).CreateFile(context.Background(), f.path,
			filestore.FileType(f.typ), f.mimetytpe, []byte(f.data))
		if err != nil {
			return err
		}
	}

	return nil
}

func basicNamespace(t *testing.T, db *gorm.DB, name, script string) tsengine.Config {
	err := testNamespace(db,
		name,
		map[string]string{
			"secret1": "mysecret1",
			"secret2": "mysecret2",
		},
		[]testFile{
			{
				path:      "/myfile",
				mimetytpe: "application/text",
				data:      "this is the content",
				typ:       filestore.FileTypeFile,
			},
			{
				path: "/myflows",
				data: "this is the content",
				typ:  filestore.FileTypeDirectory,
			},
			{
				path:      "/myflows/myflow.ts",
				mimetytpe: utils.TypeScriptMimeType,
				data:      script,
				typ:       filestore.FileTypeWorkflow,
			},
		})

	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	os.Setenv("DIREKTIV_JSENGINE_FLOWPATH", "/myflows/myflow.ts")
	d, _ := os.MkdirTemp("", "testme")
	os.Setenv("DIREKTIV_JSENGINE_BASEDIR", d)
	os.Setenv("DIREKTIV_JSENGINE_SELFCOPY", filepath.Join(d, "demo"))

	cfg := tsengine.Config{}
	env.Parse(&cfg)

	return cfg
}

func TestServerBasic(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	var initTests = []struct {
		script string
		err    error
	}{
		{"function start(state) {}", nil},
		{"", nil},
		{"function {{{}", fmt.Errorf("syntax error")},
		{`function start { var s = getSecret({ name: "secret1" }) }`, nil},
	}

	for i, tt := range initTests {
		t.Run(tt.script, func(t *testing.T) {
			cfg := basicNamespace(t, db, fmt.Sprintf("n%d", i), tt.script)
			_, err := tsengine.NewServer(cfg, db)
			if (tt.err == nil && err != nil) || (tt.err != nil && err == nil) {
				t.Fatalf("server db init failed, expected %v, got %v", tt.err, err)
			}
		})
	}
}

func TestServerSecrets(t *testing.T) {

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	var initTests = []struct {
		script string
		err    error
	}{
		{`function start { var s = getSecret({ name: "secret1" }) }`, nil},
		{
			`function start { var s = getSecret({ name: "secret1" })
				var s2 = getSecret({ name: "secret2" }) }`, nil},
		{`function start { var s = getSecret({ name: "unknown" }) }`, fmt.Errorf("")},
	}

	for i, tt := range initTests {
		t.Run(tt.script, func(t *testing.T) {
			cfg := basicNamespace(t, db, fmt.Sprintf("n%d", i), tt.script)
			_, err := tsengine.NewServer(cfg, db)
			if (tt.err == nil && err != nil) || (tt.err != nil && err == nil) {
				t.Fatalf("server db init failed, expected %v, got %v", tt.err, err)
			}
		})
	}
}

func TestServerFiles(t *testing.T) {

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}

	var initTests = []struct {
		script string
		err    error
	}{
		{`function start { var s = getFile({ name: "/myfile" }) }`, nil},
		{`function start { var s = getFile({ name: "../myfile" }) }`, nil},
		{`function start { var s = getFile({ name: "../../../../../whatever" }) }`, fmt.Errorf("not exist")},
	}

	for i, tt := range initTests {
		t.Run(tt.script, func(t *testing.T) {
			cfg := basicNamespace(t, db, fmt.Sprintf("n%d", i), tt.script)
			_, err := tsengine.NewServer(cfg, db)
			if (tt.err == nil && err != nil) || (tt.err != nil && err == nil) {
				t.Fatalf("server db init failed, expected %v, got %v", tt.err, err)
			}
		})
	}
}
func TestServerFileGetter(t *testing.T) {

	srv := startServerHttp(action)
	defer srv.stop()

	script := `
	var fn = setupFunction({
		image: "localhost:5000/hello"
	})

	const secret = getSecret({
		name: "mysecret"
	})

	function start(state) {
		var result =  fn.execute({
			input: state.data()["data"]
		})

		var r = {}
		r["secret"] = secret.string()
		r["value"] = result["return"]

		return r
	}
	`

	f, _ := os.MkdirTemp("", "test")
	scriptPath := filepath.Join(f, "myflow.ts")
	os.WriteFile(scriptPath, []byte(script), 0755)

	// add secrets
	os.Mkdir(filepath.Join(f, "secrets"), 0755)
	os.WriteFile(filepath.Join(f, "secrets", "mysecret"), []byte("secretvalue"), 0755)
	os.Setenv("DIREKTIV_JSENGINE_INIT", "file")
	os.Setenv("DIREKTIV_JSENGINE_FLOWPATH", scriptPath)
	os.Setenv("DIREKTIV_JSENGINE_BASEDIR", f)

	var fn = make(map[string]interface{})
	fn["image"] = "localhost:5000/hello"
	fnID, _ := compiler.GenerateFunctionID(fn)
	os.Setenv(fnID, fmt.Sprintf("http://127.0.0.1:%d", srv.port))

	cfg := tsengine.Config{}
	env.Parse(&cfg)

	s, _ := tsengine.NewServer(cfg, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/dummy", strings.NewReader(string("{ \"data\": \"coming-in\"}")))

	s.Engine.RunRequest(req, w)

	r, _ := io.ReadAll(w.Result().Body)
	assert.Equal(t, "{\"secret\":\"secretvalue\",\"value\":\"coming-in\"}", string(r))

}

func action(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	w.Write([]byte(fmt.Sprintf("{ \"return\": \"%s\" }", string(b))))
}

type httpServer struct {
	srv  *http.Server
	port int
}

func startServerHttp(f func(http.ResponseWriter, *http.Request)) *httpServer {

	listener, _ := net.Listen("tcp", ":0")
	port := listener.Addr().(*net.TCPAddr).Port

	l := fmt.Sprintf(":%d", port)
	listener.Close()

	mux := &http.ServeMux{}

	srv := &http.Server{Addr: l, Handler: mux}

	mux.HandleFunc("/", f)

	s := &httpServer{
		port: port,
		srv:  srv,
	}
	fmt.Printf("serving at %d\n", s.port)

	go func() {
		srv.ListenAndServe()
	}()

	return s

}

func (srv *httpServer) stop() {
	srv.srv.Shutdown(context.Background())
}
