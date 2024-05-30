package filestore_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/google/uuid"
)

func assertFileStoreCorrectRootCreation(t *testing.T, fs filestore.FileStore, ns string) {
	t.Helper()

	root, err := fs.CreateRoot(context.Background(), uuid.New(), ns)
	if err != nil {
		t.Errorf("unexpected CreateRoot() error: %v", err)

		return
	}
	if root == nil {
		t.Errorf("unexpected nil root CreateRoot()")

		return
	}
	if root.Namespace != ns {
		t.Errorf("unexpected root.Namespace, got: >%s<, want: >%s<", root.Namespace, ns)

		return
	}
}

func assertFileStoreHasRoot(t *testing.T, fs filestore.FileStore, nsList ...string) {
	t.Helper()

	all, err := fs.GetAllRoots(context.Background())
	if err != nil {
		t.Errorf("unexpected GetAllRoots() error: %v", err)

		return
	}
	if len(all) != len(nsList) {
		t.Errorf("unexpected GetAllRoots() length, got: %d, want: %d", len(all), len(nsList))

		return
	}

	for i := range nsList {
		if all[i].Namespace != nsList[i] {
			t.Errorf("unexpected all[%d].ID , got: >%s<, want: >%s<", i, all[i].Namespace, nsList[i])

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
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	tests := []struct {
		name string
		id   string
	}{
		{"validCase", "ns1"},
		{"validCase", "ns2"},
		{"validCase", "ns3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFileStoreCorrectRootCreation(t, fs, tt.id)
		})
	}
}

func Test_sqlFileStore_ListingAfterCreate(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	myNamespace1 := "ns1"
	myNamespace2 := "ns2"
	myNamespace3 := "ns3"

	// assert correct empty list.
	assertFileStoreHasRoot(t, fs)

	// create two roots:
	assertFileStoreCorrectRootCreation(t, fs, myNamespace1)
	assertFileStoreCorrectRootCreation(t, fs, myNamespace2)

	// assert existence.
	assertFileStoreHasRoot(t, fs, myNamespace1, myNamespace2)

	// add a third one:
	assertFileStoreCorrectRootCreation(t, fs, myNamespace3)

	// assert existence:
	assertFileStoreHasRoot(t, fs, myNamespace1, myNamespace2, myNamespace3)

	root, err := fs.GetRootByNamespace(context.Background(), myNamespace2)
	if err != nil {
		panic(err)
	}

	// delete one:
	assertFileStoreCorrectRootDeletion(t, fs, root.ID)

	// assert correct list:
	assertFileStoreHasRoot(t, fs, myNamespace1, myNamespace3)

	roots, err := fs.GetAllRoots(context.Background())
	if err != nil {
		panic(err)
	}
	if len(roots) != 2 {
		panic(len(roots))
	}

	// delete all:
	assertFileStoreCorrectRootDeletion(t, fs, roots[0].ID, roots[1].ID)

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
