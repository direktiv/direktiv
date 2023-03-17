package mirror_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/psql"
	"github.com/direktiv/direktiv/pkg/refactor/mirror"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestExecuteMirroringProcess(t *testing.T) {
	fs, err := psql.NewMockFileStore()
	if err != nil {
		t.Fatalf("unepxected NewMockFilestore() error = %v", err)
	}

	direktivRoot, err := fs.CreateRoot(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unepxected GetRoot() error = %v", err)
	}

	source := &mirror.MockedSource{
		Paths: map[string]string{
			"/file1.text": "file 1 content",
			"/file2.text": "file 2 content",
			"/file3.text": "file 3 content",
		},
	}

	err = mirror.ExecuteMirroringProcess(context.Background(), zap.NewNop().Sugar(),
		fs, direktivRoot, source, mirror.Settings{})
	if err != nil {
		t.Fatalf("unepxected ExecuteMirroringProcess() error = %v", err)
	}

	assertRootFilesInPath(t, fs, direktivRoot, "/",
		"/file1.text",
		"/file2.text",
		"/file3.text",
	)
}

func assertRootFilesInPath(t *testing.T, fs filestore.FileStore, root *filestore.Root, searchPath string, paths ...string) {
	t.Helper()

	files, err := fs.ForRoot(root).ReadDirectory(context.Background(), searchPath)
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
