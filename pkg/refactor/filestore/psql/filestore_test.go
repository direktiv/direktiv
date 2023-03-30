package psql_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/psql"
	"github.com/google/uuid"
)

func assertFileStoreCorrectRootCreation(t *testing.T, fs filestore.FileStore, id uuid.UUID) {
	t.Helper()

	root, err := fs.CreateRoot(context.Background(), id)
	if err != nil {
		t.Errorf("unexpected CreateRoot() error: %v", err)
	}
	if root == nil {
		t.Errorf("unexpected nil root CreateRoot()")
	}
	if root.ID != id {
		t.Errorf("unexpected root.ID, got: >%s<, want: >%s<", root.ID, id)
	}
}

func assertFileStoreHasRoot(t *testing.T, fs filestore.FileStore, ids ...uuid.UUID) {
	t.Helper()

	all, err := fs.GetAllRoots(context.Background())
	if err != nil {
		t.Errorf("unexpected GetAllRoots() error: %v", err)
	}
	if len(all) != len(ids) {
		t.Errorf("unexpected GetAllRoots() length, got: %d, want: %d", len(all), len(ids))
	}

	for i := range ids {
		if all[i].ID != ids[i] {
			t.Errorf("unexpected all[%d].ID , got: >%s<, want: >%s<", i, all[i].ID, ids[i])
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
	fs, err := psql.NewMockFileStore()
	if err != nil {
		t.Fatalf("unepxected NewMockFileStore() error = %v", err)
	}

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
	fs, err := psql.NewMockFileStore()
	if err != nil {
		t.Fatalf("create mock filestore: %s", err)
	}

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
