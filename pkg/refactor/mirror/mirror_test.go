package mirror_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/datastore"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
)

type testCallbacks struct {
	store    datastore.MirrorStore
	fstore   filestore.FileStore
	varstore datastore.RuntimeVariablesStore
	buf      *bytes.Buffer
}

func (c *testCallbacks) ConfigureWorkflowFunc(ctx context.Context, nsID uuid.UUID, nsName string, file *filestore.File) error {
	return nil
}

func (c *testCallbacks) SysLogCrit(msg string) {
}

func (c *testCallbacks) Store() datastore.MirrorStore {
	return c.store
}

func (c *testCallbacks) FileStore() filestore.FileStore {
	return c.fstore
}

func (c *testCallbacks) VarStore() datastore.RuntimeVariablesStore {
	return c.varstore
}

var _ mirror.Callbacks = &testCallbacks{}

func assertProcessSuccess(ctx context.Context, callbacks mirror.Callbacks, t *testing.T, pid uuid.UUID) {
	p, err := callbacks.Store().GetProcess(ctx, pid)
	if err != nil {
		t.Fatalf("unexpected GetProcess() error = %v", err)
	}

	if p.Status != datastore.ProcessStatusComplete {
		t.Errorf("assertProcessSuccess failed: expected %s but got %s", datastore.ProcessStatusComplete, p.Status)
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
