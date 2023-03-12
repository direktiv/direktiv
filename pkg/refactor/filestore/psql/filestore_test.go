package psql_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/psql"
	"github.com/google/uuid"
)

func assertFilestoreCorrectRootCreation(t *testing.T, fs filestore.Filestore, id uuid.UUID) {
	t.Helper()

	root, err := fs.CreateRoot(id)
	if err != nil {
		t.Errorf("unexpected CreateRoot() error: %v", err)
	}
	if root == nil {
		t.Errorf("unexpected nil root CreateRoot()")
	}
	if root.GetID() != id {
		t.Errorf("unexpected GetID(), got: >%s<, want: >%s<", root.GetID(), id)
	}
	root, err = fs.GetRoot(id)
	if err != nil {
		t.Errorf("unexpected GetRoot() error: %v", err)
	}
	if root == nil {
		t.Errorf("unexpected nil namepace")
	}
	if root.GetID() != id {
		t.Errorf("unexpected second GetID(), got: >%s<, want: >%s<", root.GetID(), id)
	}
}

func assertFilestoreHasRoot(t *testing.T, fs filestore.Filestore, ids ...uuid.UUID) {
	t.Helper()

	all, err := fs.GetAllRoots()
	if err != nil {
		t.Errorf("unexpected GetAllRoots() error: %v", err)
	}
	if len(all) != len(ids) {
		t.Errorf("unexpected GetAllRoots() length, got: %d, want: %d", len(all), len(ids))
	}

	for i := range ids {
		if all[i].GetID() != ids[i] {
			t.Errorf("unexpected all[%d].GetName() , got: >%s<, want: >%s<", i, all[i].GetID(), ids[i])
		}
	}
}

func assertFilestoreCorrectRootDeletion(t *testing.T, fs filestore.Filestore, ids ...uuid.UUID) {
	t.Helper()

	for i := range ids {
		root, err := fs.GetRoot(ids[i])
		if err != nil {
			t.Errorf("unexpected GetRoot() error: %v", err)
		}
		err = fs.ForRoot(root).Delete()
		if err != nil {
			t.Errorf("unexpected Delete() error: %v", err)
		}
	}
}

func Test_sqlFilestore_CreateRoot(t *testing.T) {
	fs, err := psql.NewMockFilestore()
	if err != nil {
		t.Fatalf("unepxected NewMockFilestore() error = %v", err)
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
			assertFilestoreCorrectRootCreation(t, fs, tt.id)
		})
	}
}

func Test_sqlFilestore_ListingAfterCreate(t *testing.T) {
	fs, err := psql.NewMockFilestore()
	if err != nil {
		t.Fatalf("create mock filestore: %s", err)
	}

	myRoot1 := uuid.New()
	myRoot2 := uuid.New()
	myRoot3 := uuid.New()

	// assert correct empty list.
	assertFilestoreHasRoot(t, fs)

	// create two roots:
	assertFilestoreCorrectRootCreation(t, fs, myRoot1)
	assertFilestoreCorrectRootCreation(t, fs, myRoot2)

	// assert existence.
	assertFilestoreHasRoot(t, fs, myRoot1, myRoot2)

	// add a third one:
	assertFilestoreCorrectRootCreation(t, fs, myRoot3)

	// assert existence:
	assertFilestoreHasRoot(t, fs, myRoot1, myRoot2, myRoot3)

	// delete one:
	assertFilestoreCorrectRootDeletion(t, fs, myRoot2)

	// assert correct list:
	assertFilestoreHasRoot(t, fs, myRoot1, myRoot3)

	// delete all:
	assertFilestoreCorrectRootDeletion(t, fs, myRoot1, myRoot3)

	// assert correct empty list.
	assertFilestoreHasRoot(t, fs)
}
