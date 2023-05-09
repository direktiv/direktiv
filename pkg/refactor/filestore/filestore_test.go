package filestore_test

import (
	"context"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore/psql"
	"github.com/direktiv/direktiv/pkg/refactor/utils"
	"github.com/google/uuid"
)

func assertFileStoreCorrectRootCreation(t *testing.T, fs filestore.FileStore, id uuid.UUID) {
	t.Helper()

	root, err := fs.CreateRoot(context.Background(), id)
	if err != nil {
		t.Errorf("unexpected CreateRoot() error: %v", err)

		return
	}
	if root == nil {
		t.Errorf("unexpected nil root CreateRoot()")

		return
	}
	if root.ID != id {
		t.Errorf("unexpected root.ID, got: >%s<, want: >%s<", root.ID, id)

		return
	}
}

func assertFileStoreHasRoot(t *testing.T, fs filestore.FileStore, ids ...uuid.UUID) {
	t.Helper()

	all, err := fs.GetAllRoots(context.Background())
	if err != nil {
		t.Errorf("unexpected GetAllRoots() error: %v", err)

		return
	}
	if len(all) != len(ids) {
		t.Errorf("unexpected GetAllRoots() length, got: %d, want: %d", len(all), len(ids))

		return
	}

	for i := range ids {
		if all[i].ID != ids[i] {
			t.Errorf("unexpected all[%d].ID , got: >%s<, want: >%s<", i, all[i].ID, ids[i])

			return
		}
	}
}

func assertFileStoreCorrectRootDeletion(t *testing.T, fs filestore.FileStore, ids ...uuid.UUID) {
	t.Helper()

	for i := range ids {
		err := fs.ForRootID(ids[i]).Delete(context.Background())
		if err != nil {
			t.Errorf("unexpected Delete() error: %v", err)
		}
	}
}

func Test_sqlFileStore_CreateRoot(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := psql.NewSQLFileStore(db)

	tests := []struct {
		name string
		id   uuid.UUID
	}{
		{"validCase", uuid.New()},
		{"validCase", uuid.New()},
		{"validCase", uuid.New()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFileStoreCorrectRootCreation(t, fs, tt.id)
		})
	}
}

func Test_sqlFileStore_ListingAfterCreate(t *testing.T) {
	db, err := utils.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := psql.NewSQLFileStore(db)

	myRoot1 := uuid.New()
	myRoot2 := uuid.New()
	myRoot3 := uuid.New()

	// assert correct empty list.
	assertFileStoreHasRoot(t, fs)

	// create two roots:
	assertFileStoreCorrectRootCreation(t, fs, myRoot1)
	assertFileStoreCorrectRootCreation(t, fs, myRoot2)

	// assert existence.
	assertFileStoreHasRoot(t, fs, myRoot1, myRoot2)

	// add a third one:
	assertFileStoreCorrectRootCreation(t, fs, myRoot3)

	// assert existence:
	assertFileStoreHasRoot(t, fs, myRoot1, myRoot2, myRoot3)

	// delete one:
	assertFileStoreCorrectRootDeletion(t, fs, myRoot2)

	// assert correct list:
	assertFileStoreHasRoot(t, fs, myRoot1, myRoot3)

	// delete all:
	assertFileStoreCorrectRootDeletion(t, fs, myRoot1, myRoot3)

	// assert correct empty list.
	assertFileStoreHasRoot(t, fs)
}

func TestSha256CalculateChecksum(t *testing.T) {
	got := string(filestore.Sha256CalculateChecksum([]byte("some_string")))
	want := "539a374ff43dce2e894fd4061aa545e6f7f5972d40ee9a1676901fb92125ffee"
	if got != want {
		t.Errorf("unexpected Sha256CalculateChecksum() result, got: %s, want: %s", got, want)
	}
}
