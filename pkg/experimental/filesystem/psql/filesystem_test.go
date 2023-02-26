package psql_test

import (
	"context"
	"github.com/direktiv/direktiv/pkg/experimental/filesystem"
	"github.com/direktiv/direktiv/pkg/experimental/filesystem/psql"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"testing"
)

func createMockFilesystem() (filesystem.Filesystem, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&psql.Namespace{}, &psql.File{})
	if err != nil {
		return nil, err
	}
	fs := psql.NewSqlFilesystem(db)

	return fs, nil
}

func assertFilesystemCorrectNamespaceCreation(t *testing.T, fs filesystem.Filesystem, namespace string) {
	ns, err := fs.CreateNamespace(context.Background(), namespace)
	if err != nil {
		t.Errorf("unexpected CreateNamespace() error: %v", err)
	}
	if ns == nil {
		t.Errorf("unexpected nil namepace CreateNamespace()")
	}
	if ns.GetName() != namespace {
		t.Errorf("unexpected GetName(), got: >%s<, want: >%s<", ns.GetName(), namespace)
	}
	ns, err = fs.GetNamespace(context.Background(), namespace)
	if err != nil {
		t.Errorf("unexpected GetNamespace() error: %v", err)
	}
	if ns == nil {
		t.Errorf("unexpected nil namepace")
	}
	if ns.GetName() != namespace {
		t.Errorf("unexpected second GetName(), got: >%s<, want: >%s<", ns.GetName(), namespace)
	}
}

func assertFilesystemHasNamespace(t *testing.T, fs filesystem.Filesystem, names ...string) {
	all, err := fs.GetAllNamespaces(context.Background())
	if err != nil {
		t.Errorf("unexpected GetAllNamespaces() error: %v", err)
	}
	if len(all) != len(names) {
		t.Errorf("unexpected GetAllNamespaces() length, got: %d, want: %d", len(all), len(names))
	}

	for i, _ := range names {
		if all[i].GetName() != names[i] {
			t.Errorf("unexpected all[%d].GetName() , got: >%s<, want: >%s<", i, all[i].GetName(), names[i])
		}
	}
}

func assertFilesystemCorrectNamespaceDeletion(t *testing.T, fs filesystem.Filesystem, names ...string) {
	for i, _ := range names {
		err := fs.DeleteNamespace(context.Background(), names[i])
		if err != nil {
			t.Errorf("unexpected DeleteNamespace() error: %v", err)
		}
	}
}

func Test_sqlFilesystem_CreateNamespace(t *testing.T) {
	fs, err := createMockFilesystem()
	if err != nil {
		t.Fatalf("unepxected createMockFilesystem() error = %v", err)
	}

	tests := []struct {
		name      string
		namespace string
	}{
		{"validCase", "namespace_1"},
		{"validCase", "namespace_2"},
		{"validCase", "namespace_3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertFilesystemCorrectNamespaceCreation(t, fs, tt.namespace)
		})
	}
}

func Test_sqlFilesystem_ListingAfterCreate(t *testing.T) {
	fs, err := createMockFilesystem()
	if err != nil {
		t.Fatalf("create mock filesystem: %s", err)
	}

	// assert correct empty list.
	assertFilesystemHasNamespace(t, fs)

	// create two namespaces:
	assertFilesystemCorrectNamespaceCreation(t, fs, "my_namespace_1")
	assertFilesystemCorrectNamespaceCreation(t, fs, "my_namespace_2")

	// assert existence.
	assertFilesystemHasNamespace(t, fs, "my_namespace_1", "my_namespace_2")

	// add a third one:
	assertFilesystemCorrectNamespaceCreation(t, fs, "my_namespace_3")

	// assert existence:
	assertFilesystemHasNamespace(t, fs, "my_namespace_1", "my_namespace_2", "my_namespace_3")

	// delete one:
	assertFilesystemCorrectNamespaceDeletion(t, fs, "my_namespace_2")

	// assert correct list:
	assertFilesystemHasNamespace(t, fs, "my_namespace_1", "my_namespace_3")

	// delete all:
	assertFilesystemCorrectNamespaceDeletion(t, fs, "my_namespace_1", "my_namespace_3")

	// assert correct empty list.
	assertFilesystemHasNamespace(t, fs)
}
