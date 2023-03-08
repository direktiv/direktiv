package psql_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/psql"
	"github.com/google/uuid"
)

func TestRoot_CreateFile(t *testing.T) {
	fs, err := psql.NewMockFilestore()
	if err != nil {
		t.Fatalf("unepxected NewMockFilestore() error = %v", err)
	}
	root, err := fs.CreateRoot(context.Background(), uuid.UUID{})
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
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
			assertRootCorrectFileCreation(t, root, tt.path, tt.typ, []byte(tt.payload))
		})
	}
}

func assertRootCorrectFileCreation(t *testing.T, ns filestore.Root, path string, typ string, data []byte) {
	t.Helper()

	file, err := ns.CreateFile(context.Background(), path, filestore.FileType(typ), bytes.NewReader(data))
	if err != nil {
		t.Errorf("unexpected CreateFile() error: %v", err)
	}
	if file == nil {
		t.Errorf("unexpected nil file CreateFile()")
	}
	if file.GetPath() != path {
		t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", file.GetPath(), path)
	}
	reader, _ := file.GetData(context.Background())
	createdData, _ := io.ReadAll(reader)
	if string(createdData) != string(data) {
		t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", createdData, data)
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

func TestRoot_CorrectListPath(t *testing.T) {
	fs, err := psql.NewMockFilestore()
	if err != nil {
		t.Fatalf("unepxected NewMockFilestore() error = %v", err)
	}
	root, err := fs.CreateRoot(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	// Test root directory:
	{
		assertRootCorrectFileCreation(t, root, "/file1.text", "text", []byte("content1"))
		assertRootCorrectFileCreation(t, root, "/file2.text", "text", []byte("content2"))

		assertRootFilesInPath(t, root, "/",
			"/file1.text",
			"/file2.text",
		)
	}

	// Add /dir1 directory:
	{
		assertRootCorrectFileCreation(t, root, "/dir1", "directory", nil)
		assertRootCorrectFileCreation(t, root, "/dir1/file3.text", "text", []byte("content3"))
		assertRootCorrectFileCreation(t, root, "/dir1/file4.text", "text", []byte("content4"))

		assertRootFilesInPath(t, root, "/dir1",
			"/dir1/file3.text",
			"/dir1/file4.text",
		)
		assertRootFilesInPath(t, root, "/",
			"/file1.text",
			"/file2.text",
			"/dir1",
		)
	}

	// Add /dir1/dir2 directory:
	{
		assertRootCorrectFileCreation(t, root, "/dir1/dir2", "directory", nil)
		assertRootCorrectFileCreation(t, root, "/dir1/dir2/file5.text", "text", []byte("content5"))
		assertRootCorrectFileCreation(t, root, "/dir1/dir2/file6.text", "text", []byte("content6"))

		assertRootFilesInPath(t, root, "/dir1/dir2",
			"/dir1/dir2/file5.text",
			"/dir1/dir2/file6.text",
		)
		assertRootFilesInPath(t, root, "/dir1",
			"/dir1/file3.text",
			"/dir1/file4.text",
			"/dir1/dir2",
		)
		assertRootFilesInPath(t, root, "/",
			"/file1.text",
			"/file2.text",
			"/dir1",
		)
	}
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
