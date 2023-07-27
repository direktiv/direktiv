package mirror_test

import (
	"context"
	"testing"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/datastoresql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
)

func TestExecuteMirroringProcess(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	dStore := datastoresql.NewSQLStore(db, "some_secret_key_")
	store := dStore.Mirror()

	direktivRoot, err := fs.CreateRoot(context.Background(), uuid.New(), "test")
	if err != nil {
		t.Fatalf("unepxected GetRoot() error = %v", err)
	}

	config, err := store.CreateConfig(context.Background(), &mirror.Config{
		NamespaceID: direktivRoot.ID,
	})
	if err != nil {
		t.Fatalf("unepxected CreateConfig() error = %v", err)
	}

	source := &mirror.MockedSource{
		Paths: map[string]string{
			"/file1.text": "file 1 content",
			"/file2.text": "file 2 content",
			"/file3.text": "file 3 content",
		},
	}

	manager := mirror.NewDefaultManager(nil, nil, store, fs, dStore.RuntimeVariables(), source, nil)

	_, err = manager.StartInitialMirroringProcess(context.Background(), config)
	if err != nil {
		t.Fatalf("unepxected ExecuteMirroringProcess() error = %v", err)
	}
	time.Sleep(time.Second)

	assertRootFilesInPath(t, fs, direktivRoot, "/",
		"/file1.text",
		"/file2.text",
		"/file3.text",
	)
}

func assertRootFilesInPath(t *testing.T, fs filestore.FileStore, root *filestore.Root, searchPath string, paths ...string) {
	t.Helper()

	files, err := fs.ForRootID(root.ID).ReadDirectory(context.Background(), searchPath)
	if err != nil {
		t.Errorf("unepxected ReadDirectory() error = %v", err)
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
