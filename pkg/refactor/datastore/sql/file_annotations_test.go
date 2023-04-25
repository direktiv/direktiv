package sql_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/datastore/sql"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/psql"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

func Test_sqlFileAnnotationsStore_SetAndGet(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	ds := sql.NewSQLStore(db, "some_secret_key_")
	fs := psql.NewSQLFileStore(db)

	file := createFile(t, fs)

	err = ds.FileAnnotations().Set(context.Background(), &core.FileAnnotations{
		FileID: file.ID,
		Data: map[string]string{
			"some_key": "some_data",
		},
	})
	if err != nil {
		t.Errorf("unexpected Set() error: %v", err)

		return
	}

	annotations, err := ds.FileAnnotations().Get(context.Background(), file.ID)
	if err != nil {
		t.Errorf("unexpected Get() error: %v", err)

		return
	}

	if annotations.FileID != file.ID {
		t.Errorf("unexpected Get().ID, got %s, want %s", annotations.FileID, file.ID)

		return
	}

	wantData := core.FileAnnotationsData{
		"some_key": "some_data",
	}
	if !reflect.DeepEqual(wantData, annotations.Data) {
		t.Errorf("unexpected Get().Data, want %s, got %s", wantData, annotations.Data)

		return
	}
}

func createFile(t *testing.T, fs filestore.FileStore) *filestore.File {
	t.Helper()

	id := uuid.New()

	_, err := fs.CreateRoot(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected CreateRoot() error: %v", err)
	}
	_, _, err = fs.ForRootID(id).CreateFile(context.Background(), "/", filestore.FileTypeDirectory, nil)
	if err != nil {
		t.Fatalf("unexpected CreateFile() error: %v", err)
	}

	file, _, err := fs.ForRootID(id).CreateFile(context.Background(), "/my_file.text", filestore.FileTypeFile, strings.NewReader("my file"))
	if err != nil {
		t.Fatalf("unexpected CreateFile() error: %v", err)
	}

	return file
}
