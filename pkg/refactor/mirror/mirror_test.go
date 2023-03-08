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
	fs, err := psql.NewMockFilestore()
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
		direktivRoot, source, mirror.Settings{})
	if err != nil {
		t.Fatalf("unepxected ExecuteMirroringProcess() error = %v", err)
	}

	assertRootFilesInPath(t, direktivRoot, "/",
		"/file1.text",
		"/file2.text",
		"/file3.text",
	)
}

func assertRootFilesInPath(t *testing.T, root filestore.Root, searchPath string, paths ...string) {
	t.Helper()

	files, err := root.ListPath(context.Background(), searchPath)
	if err != nil {
		t.Errorf("unepxected ListPath() error = %v", err)
	}
	if len(files) != len(paths) {
		t.Errorf("unexpected ListPath() length, got: %d, want: %d", len(files), len(paths))
	}

	for i := range paths {
		if files[i].GetPath() != paths[i] {
			t.Errorf("unexpected files[%d].GetPath() , got: >%s<, want: >%s<", i, files[i].GetPath(), paths[i])
		}
	}
}
