package sql_test

import (
	"context"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/sql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/psql"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

func Test_sqlFileAttributesStore_SetAndGet(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	ds := sql.NewSQLStore(db)
	fs := psql.NewSQLFileStore(db)

	file := createFile(t, fs)

	err = ds.FileAttributes().Set(context.Background(), &core.FileAttributes{
		FileID: file.ID,
		Value:  "some attributes",
	})
	if err != nil {
		t.Errorf("unexpected Set() error: %v", err)
	}

	attrs, err := ds.FileAttributes().Get(context.Background(), file.ID)
	if err != nil {
		t.Errorf("unexpected Get() error: %v", err)
	}

	if attrs.FileID != file.ID {
		t.Errorf("unexpected Get().ID, got %s, want %s", attrs.FileID, file.ID)
	}

	wantValue := "some attributes"
	if string(attrs.Value) != wantValue {
		t.Errorf("unexpected Get().Value, want %s, got %s", wantValue, attrs.Value)
	}
}

func createFile(t *testing.T, fs filestore.FileStore) *filestore.File {
	t.Helper()

	id := uuid.New()

	_, err := fs.CreateRoot(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected CreateRoot() error: %v", err)
	}

	file, _, err := fs.ForRootID(id).CreateFile(context.Background(), "/my_file.text", filestore.FileTypeFile, strings.NewReader("my file"))
	if err != nil {
		t.Fatalf("unexpected CreateFile() error: %v", err)
	}
	return file
}
