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
	fs, err := psql.NewMockFileStore()
	if err != nil {
		t.Fatalf("unepxected NewMockFileStore() error = %v", err)
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
			assertRootCorrectFileCreation(t, fs, root, tt.path, tt.typ, []byte(tt.payload))
		})
	}
}

func assertRootCorrectFileCreation(t *testing.T, fs filestore.FileStore, root filestore.Root, path string, typ string, data []byte) {
	t.Helper()

	file, err := fs.ForRoot(root).CreateFile(context.Background(), path, filestore.FileType(typ), bytes.NewReader(data))
	if err != nil {
		t.Errorf("unexpected CreateFile() error: %v", err)
	}
	if file == nil {
		t.Errorf("unexpected nil file CreateFile()")
	}
	if file.GetPath() != path {
		t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", file.GetPath(), path)
	}

	if typ != "directory" {
		reader, _ := fs.ForFile(file).GetData(context.Background())
		createdData, _ := io.ReadAll(reader)
		if string(createdData) != string(data) {
			t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", createdData, data)
		}
	}

	file, err = fs.ForRoot(root).GetFile(context.Background(), path, nil)
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

func TestRoot_CorrectReadDirectory(t *testing.T) {
	fs, err := psql.NewMockFileStore()
	if err != nil {
		t.Fatalf("unepxected NewMockFileStore() error = %v", err)
	}
	root, err := fs.CreateRoot(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	// Test root directory:
	{
		assertRootCorrectFileCreation(t, fs, root, "/file1.text", "text", []byte("content1"))
		assertRootCorrectFileCreation(t, fs, root, "/file2.text", "text", []byte("content2"))

		assertRootFilesInPath(t, fs, root, "/",
			"/file1.text",
			"/file2.text",
		)
	}

	// Add /dir1 directory:
	{
		assertRootCorrectFileCreation(t, fs, root, "/dir1", "directory", nil)
		assertRootCorrectFileCreation(t, fs, root, "/dir1/file3.text", "text", []byte("content3"))
		assertRootCorrectFileCreation(t, fs, root, "/dir1/file4.text", "text", []byte("content4"))

		assertRootFilesInPath(t, fs, root, "/dir1",
			"/dir1/file3.text",
			"/dir1/file4.text",
		)
		assertRootFilesInPath(t, fs, root, "/",
			"/file1.text",
			"/file2.text",
			"/dir1",
		)
	}

	// Add /dir1/dir2 directory:
	{
		assertRootCorrectFileCreation(t, fs, root, "/dir1/dir2", "directory", nil)
		assertRootCorrectFileCreation(t, fs, root, "/dir1/dir2/file5.text", "text", []byte("content5"))
		assertRootCorrectFileCreation(t, fs, root, "/dir1/dir2/file6.text", "text", []byte("content6"))

		assertRootFilesInPath(t, fs, root, "/dir1/dir2",
			"/dir1/dir2/file5.text",
			"/dir1/dir2/file6.text",
		)
		assertRootFilesInPath(t, fs, root, "/dir1",
			"/dir1/file3.text",
			"/dir1/file4.text",
			"/dir1/dir2",
		)
		assertRootFilesInPath(t, fs, root, "/",
			"/file1.text",
			"/file2.text",
			"/dir1",
		)
	}
}

func assertRootFilesInPath(t *testing.T, fs filestore.FileStore, root filestore.Root, searchPath string, paths ...string) {
	t.Helper()

	files, err := fs.ForRoot(root).ReadDirectory(context.Background(), searchPath)
	if err != nil {
		t.Errorf("unepxected ReadDirectory() error = %v", err)
	}
	if len(files) != len(paths) {
		t.Errorf("unexpected ReadDirectory() length, got: %d, want: %d", len(files), len(paths))
	}

	for i := range paths {
		if files[i].GetPath() != paths[i] {
			t.Errorf("unexpected files[%d].GetPath() , got: >%s<, want: >%s<", i, files[i].GetPath(), paths[i])
		}
	}
}
