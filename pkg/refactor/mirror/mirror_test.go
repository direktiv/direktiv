package mirror_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
)

type testLogger struct {
	buf *bytes.Buffer
}

func (l *testLogger) Error(processID uuid.UUID, msg string, keysAndValues ...interface{}) {
	l.buf.WriteString(fmt.Sprintf("EROR %s\n", msg))
}

func (l *testLogger) Warn(processID uuid.UUID, msg string, keysAndValues ...interface{}) {
	l.buf.WriteString(fmt.Sprintf("WARN %s\n", msg))
}

func (l *testLogger) Info(processID uuid.UUID, msg string, keysAndValues ...interface{}) {
	l.buf.WriteString(fmt.Sprintf("INFO %s\n", msg))
}

func (l *testLogger) Debug(processID uuid.UUID, msg string, keysAndValues ...interface{}) {
	// l.buf.WriteString(fmt.Sprintf("DBUG %s\n", msg))
}

var _ mirror.ProcessLogger = &testLogger{}

type testCallbacks struct {
	store    mirror.Store
	fstore   filestore.FileStore
	varstore core.RuntimeVariablesStore
	buf      *bytes.Buffer
}

func (c *testCallbacks) ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
	return nil
}

func (c *testCallbacks) ProcessLogger() mirror.ProcessLogger {
	return &testLogger{buf: c.buf}
}

func (c *testCallbacks) SysLogCrit(msg string) {
}

func (c *testCallbacks) Store() mirror.Store {
	return c.store
}

func (c *testCallbacks) FileStore() filestore.FileStore {
	return c.fstore
}

func (c *testCallbacks) VarStore() core.RuntimeVariablesStore {
	return c.varstore
}

var _ mirror.Callbacks = &testCallbacks{}

func assertProcessSuccess(ctx context.Context, callbacks mirror.Callbacks, t *testing.T, pid uuid.UUID) {
	p, err := callbacks.Store().GetProcess(ctx, pid)
	if err != nil {
		t.Fatalf("unexpected GetProcess() error = %v", err)
	}

	if p.Status != mirror.ProcessStatusComplete {
		t.Errorf("assertProcessSuccess failed: expected %s but got %s", mirror.ProcessStatusComplete, p.Status)
	}
}

func assertRootFilesInPath(t *testing.T, fs filestore.FileStore, root *filestore.Root, paths ...string) {
	t.Helper()

	files, err := fs.ForRootID(root.ID).ListAllFiles(context.Background())
	if err != nil {
		t.Errorf("unexpected ReadDirectory() error = %v", err)
	}
	if len(files) != len(paths) {
		t.Errorf("unexpected ReadDirectory() length, got: %d, want: %d", len(files), len(paths))
	}

	for i := range paths {
		if files[i].Path != paths[i] {
			t.Errorf("unexpected files[%d].Path , got: >%s<, want: >%s<", i, files[i].Path, paths[i])
		}
	}
}

func TestDryRun(t *testing.T) {
	ctx := context.Background()

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	dStore := datastoresql.NewSQLStore(db, "some_secret_key_")
	store := dStore.Mirror()

	callbacks := &testCallbacks{
		store:    store,
		fstore:   fs,
		varstore: dStore.RuntimeVariables(),
		buf:      new(bytes.Buffer),
	}

	ns := &core.Namespace{
		ID:   uuid.New(),
		Name: uuid.New().String(),
	}
	rootID := uuid.New()
	direktivRoot, err := fs.CreateRoot(ctx, rootID, ns.Name)
	if err != nil {
		t.Fatalf("unexpected GetRoot() error = %v", err)
	}
	if direktivRoot.ID != rootID {
		t.Fatal("Got wrong id back")
	}
	/*
		config, err := store.CreateConfig(ctx, &mirror.Config{
			NamespaceID: nsID,
			RootName:    "test",
		})
		if err != nil {
			t.Fatalf("unexpected CreateConfig() error = %v", err)
		}
	*/

	src := newMemSource()
	_ = src.fs.WriteFile(".direktivignore", []byte(``), 0o755)
	_ = src.fs.WriteFile("x.yaml", []byte(`x: 5`), 0o755)
	_ = src.fs.WriteFile("y.json", []byte(`{}`), 0o755)
	_ = src.fs.MkdirAll("a/b", 0o755)
	_ = src.fs.WriteFile("a/b/c.yaml", []byte(`direktiv_api: workflow/v1
states:
- id: a
  type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/d.yaml", []byte(`
states:
- id: a
	type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/e.yaml", []byte(`
states:
- id: a
  type: noop
`), 0o755)

	manager := mirror.NewManager(callbacks)

	p, err := manager.NewProcess(ctx, ns, mirror.ProcessTypeDryRun)
	if err != nil {
		t.Fatalf("unexpected NewProcess() error = %v", err)
	}

	manager.Execute(ctx, p, func(ctx context.Context) (mirror.Source, error) { return src, nil }, &mirror.DryrunApplyer{})

	assertProcessSuccess(ctx, callbacks, t, p.ID)

	root, err := callbacks.FileStore().GetRootByNamespace(ctx, ns.Name)
	if err != nil {
		t.Fatalf("unexpected GetAllRootsForNamespace() error = %v", err)
	}
	assertRootFilesInPath(t, fs, root)
}

func TestInitSync(t *testing.T) {
	ctx := context.Background()

	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unexpected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	dStore := datastoresql.NewSQLStore(db, "some_secret_key_")
	store := dStore.Mirror()

	callbacks := &testCallbacks{
		store:    store,
		fstore:   fs,
		varstore: dStore.RuntimeVariables(),
		buf:      new(bytes.Buffer),
	}

	ns := &core.Namespace{
		ID:   uuid.New(),
		Name: uuid.New().String(),
	}
	rootID := uuid.New()
	direktivRoot, err := fs.CreateRoot(ctx, rootID, ns.Name)
	if err != nil {
		t.Fatalf("unexpected GetRoot() error = %v", err)
	}
	if direktivRoot.ID != rootID {
		t.Fatal("Got wrong id back")
	}

	_, err = store.CreateConfig(ctx, &mirror.Config{
		Namespace: ns.Name,
	})
	if err != nil {
		t.Fatalf("unexpected CreateConfig() error = %v", err)
	}

	src := newMemSource()
	_ = src.fs.WriteFile(".direktivignore", []byte(``), 0o755)
	_ = src.fs.WriteFile("x.yaml", []byte(`x: 5`), 0o755)
	_ = src.fs.WriteFile("y.json", []byte(`{}`), 0o755)
	_ = src.fs.MkdirAll("a/b", 0o755)
	_ = src.fs.WriteFile("a/b/c.yaml", []byte(`direktiv_api: workflow/v1
states:
- id: a
  type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/d.yaml", []byte(`
states:
- id: a
	type: noop
`), 0o755)
	_ = src.fs.WriteFile("a/b/e.yaml", []byte(`
states:
- id: a
  type: noop
`), 0o755)

	manager := mirror.NewManager(callbacks)

	p, err := manager.NewProcess(ctx, ns, mirror.ProcessTypeInit)
	if err != nil {
		t.Fatalf("unexpected NewProcess() error = %v", err)
	}

	manager.Execute(ctx, p, func(ctx context.Context) (mirror.Source, error) { return src, nil }, &mirror.DirektivApplyer{NamespaceID: ns.ID})

	assertProcessSuccess(ctx, callbacks, t, p.ID)

	root, err := callbacks.FileStore().GetRootByNamespace(ctx, ns.Name)
	if err != nil {
		t.Fatalf("unexpected GetAllRootsForNamespace() error = %v", err)
	}

	assertRootFilesInPath(t, fs, root,
		"/a",
		"/a/b",
		"/a/b/c.yaml",
		"/a/b/d.yaml",
		"/a/b/e.yaml",
		"/x.yaml",
		"/y.json",
	)
}
