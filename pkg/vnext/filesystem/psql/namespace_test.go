package psql_test

import (
	"context"
	"github.com/direktiv/direktiv/pkg/vnext/filesystem"
	"testing"
)

func TestNamespace_CreateFile(t *testing.T) {
	fs, err := createMockFilesystem()
	if err != nil {
		t.Fatalf("unepxected createMockFilesystem() error = %v", err)
	}
	ns, err := fs.CreateNamespace(context.Background(), "my_namespace")
	if err != nil {
		t.Fatalf("unepxected CreateNamespace() error = %v", err)
	}

	tests := []struct {
		path    string
		typ     string
		payload string
	}{
		{"/example.text", "text", "abcd"},
		{"/example.text", "text", "abcd"},
		{"/example.text", "text", "abcd"},
	}
	for _, tt := range tests {
		t.Run("valid", func(t *testing.T) {
			assertNamespaceCorrectFileCreation(t, ns, tt.path, tt.typ, []byte(tt.payload))
		})
	}
}

func assertNamespaceCorrectFileCreation(t *testing.T, ns filesystem.Namespace, path string, typ string, payload []byte) {
	file, err := ns.CreateFile(context.Background(), path, typ, payload)
	if err != nil {
		t.Errorf("unexpected CreateFile() error: %v", err)
	}
	if file == nil {
		t.Errorf("unexpected nil file CreateFile()")
	}
	if file.GetPath() != path {
		t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", file.GetPath(), path)
	}
	if string(file.GetPayload()) != string(payload) {
		t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", file.GetPayload(), payload)
	}
	file, err = ns.GetFile(context.Background(), path)
	if err != nil {
		t.Errorf("unexpected GetFile() error: %v", err)
	}
	if file == nil {
		t.Errorf("unexpected nil file GetFile()")
	}
	if file.GetPath() != path {
		t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", file.GetPath(), path)
	}
}

func TestNamespace_CorrectListPath(t *testing.T) {
	fs, err := createMockFilesystem()
	if err != nil {
		t.Fatalf("unepxected createMockFilesystem() error = %v", err)
	}
	ns, err := fs.CreateNamespace(context.Background(), "my_namespace")
	if err != nil {
		t.Fatalf("unepxected CreateNamespace() error = %v", err)
	}

	// Test root directory:
	{
		assertNamespaceCorrectFileCreation(t, ns, "/file1.text", "text", []byte("content1"))
		assertNamespaceCorrectFileCreation(t, ns, "/file2.text", "text", []byte("content2"))

		assertNamespaceFilesInPath(t, ns, "/",
			"/file1.text",
			"/file2.text",
		)
	}

	// Add /dir1 directory:
	{
		assertNamespaceCorrectFileCreation(t, ns, "/dir1", "directory", nil)
		assertNamespaceCorrectFileCreation(t, ns, "/dir1/file3.text", "text", []byte("content3"))
		assertNamespaceCorrectFileCreation(t, ns, "/dir1/file4.text", "text", []byte("content4"))

		assertNamespaceFilesInPath(t, ns, "/dir1",
			"/dir1/file3.text",
			"/dir1/file4.text",
		)
		assertNamespaceFilesInPath(t, ns, "/",
			"/file1.text",
			"/file2.text",
			"/dir1",
		)
	}

	// Add /dir1/dir2 directory:
	{
		assertNamespaceCorrectFileCreation(t, ns, "/dir1/dir2", "directory", nil)
		assertNamespaceCorrectFileCreation(t, ns, "/dir1/dir2/file5.text", "text", []byte("content5"))
		assertNamespaceCorrectFileCreation(t, ns, "/dir1/dir2/file6.text", "text", []byte("content6"))

		assertNamespaceFilesInPath(t, ns, "/dir1/dir2",
			"/dir1/dir2/file5.text",
			"/dir1/dir2/file6.text",
		)
		assertNamespaceFilesInPath(t, ns, "/dir1",
			"/dir1/file3.text",
			"/dir1/file4.text",
			"/dir1/dir2",
		)
		assertNamespaceFilesInPath(t, ns, "/",
			"/file1.text",
			"/file2.text",
			"/dir1",
		)
	}
}

func assertNamespaceFilesInPath(t *testing.T, ns filesystem.Namespace, searchPath string, paths ...string) {
	files, err := ns.ListPath(context.Background(), searchPath)
	if err != nil {
		t.Errorf("unepxected ListPath() error = %v", err)
	}
	if len(files) != len(paths) {
		t.Errorf("unexpected ListPath() length, got: %d, want: %d", len(files), len(paths))
	}
	for i, _ := range paths {
		if files[i].GetPath() != paths[i] {
			t.Errorf("unexpected files[%d].GetPath() , got: >%s<, want: >%s<", i, files[i].GetPath(), paths[i])
		}
	}
}
