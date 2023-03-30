package psql_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"reflect"
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
		{"/example1.text", "text", "abcd"},
		{"/example2.text", "text", "abcd"},
	}
	for _, tt := range tests {
		t.Run("valid", func(t *testing.T) {
			assertRootCorrectFileCreation(t, fs, root, tt.path, tt.typ, []byte(tt.payload))
		})
	}
}

func assertRootCorrectFileCreation(t *testing.T, fs filestore.FileStore, root *filestore.Root, path string, typ string, data []byte) {
	t.Helper()

	file, _, err := fs.ForRootID(root.ID).CreateFile(context.Background(), path, filestore.FileType(typ), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected CreateFile() error: %v", err)
	}
	if file == nil {
		t.Fatalf("unexpected nil file CreateFile()")
	}
	if file.Path != path {
		t.Fatalf("unexpected file.Path, got: >%s<, want: >%s<", file.Path, path)
	}

	if typ != "directory" {
		reader, _ := fs.ForFile(file).GetData(context.Background())
		createdData, _ := io.ReadAll(reader)
		if string(createdData) != string(data) {
			t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", createdData, data)
		}
	}

	file, err = fs.ForRootID(root.ID).GetFile(context.Background(), path)
	if err != nil {
		t.Errorf("unexpected GetFile() error: %v", err)
	}
	if file == nil {
		t.Errorf("unexpected nil file GetFile()")
	}
	if file.Path != path {
		t.Errorf("unexpected file.Path, got: >%s<, want: >%s<", file.Path, path)
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

func TestRoot_CalculateChecksumDirectory(t *testing.T) {
	fs, err := psql.NewMockFileStore()
	if err != nil {
		t.Fatalf("unepxected NewMockFileStore() error = %v", err)
	}
	root, err := fs.CreateRoot(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	filestore.DefaultCalculateChecksum = func(data []byte) []byte {
		return []byte(fmt.Sprintf("---%s---", data))
	}

	// Test root directory:
	{
		assertRootCorrectFileCreation(t, fs, root, "/file1.text", "text", []byte("content1"))
		assertRootCorrectFileCreation(t, fs, root, "/file2.text", "text", []byte("content2"))

		assertChecksumsInPath(t, fs, root, "/",
			"/file1.text", "---content1---",
			"/file2.text", "---content2---",
		)
	}

	// Add /dir1 directory:
	{
		assertRootCorrectFileCreation(t, fs, root, "/dir1", "directory", nil)
		assertRootCorrectFileCreation(t, fs, root, "/dir1/file3.text", "text", []byte("content3"))
		assertRootCorrectFileCreation(t, fs, root, "/dir1/file4.text", "text", []byte("content4"))

		assertChecksumsInPath(t, fs, root, "/dir1",
			"/dir1/file3.text", "---content3---",
			"/dir1/file4.text", "---content4---",
		)
		assertChecksumsInPath(t, fs, root, "/",
			"/file1.text", "---content1---",
			"/file2.text", "---content2---",
			"/dir1", "",
		)
	}

	// Add /dir1/dir2 directory:
	{
		assertRootCorrectFileCreation(t, fs, root, "/dir1/dir2", "directory", nil)
		assertRootCorrectFileCreation(t, fs, root, "/dir1/dir2/file5.text", "text", []byte("content5"))
		assertRootCorrectFileCreation(t, fs, root, "/dir1/dir2/file6.text", "text", []byte("content6"))

		assertChecksumsInPath(t, fs, root, "/dir1/dir2",
			"/dir1/dir2/file5.text", "---content5---",
			"/dir1/dir2/file6.text", "---content6---",
		)
		assertChecksumsInPath(t, fs, root, "/dir1",
			"/dir1/file3.text", "---content3---",
			"/dir1/file4.text", "---content4---",
			"/dir1/dir2", "",
		)
		assertChecksumsInPath(t, fs, root, "/",
			"/file1.text", "---content1---",
			"/file2.text", "---content2---",
			"/dir1", "",
		)
	}
}

func assertRootFilesInPath(t *testing.T, fs filestore.FileStore, root *filestore.Root, searchPath string, paths ...string) {
	t.Helper()

	files, err := fs.ForRootID(root.ID).ReadDirectory(context.Background(), searchPath)
	if err != nil {
		t.Errorf("unepxected ReadDirectory() error = %v", err)
	}
	if len(files) != len(paths) {
		t.Errorf("unexpected ReadDirectory() length, got: %d, want: %d", len(files), len(paths))
	}

	for i := range paths {
		if files[i].Path != paths[i] {
			t.Errorf("unexpected files[%d].Path , got: >%s<, want: >%s<", i, files[i].Path, paths[i])
		}
	}
}

func assertChecksumsInPath(t *testing.T, fs filestore.FileStore, root *filestore.Root, searchPath string, paths ...string) {
	t.Helper()

	checksumsMap, err := fs.ForRootID(root.ID).CalculateChecksumsMap(context.Background(), searchPath)
	if err != nil {
		t.Errorf("unepxected CalculateChecksumsMap() error = %v", err)
	}
	if len(checksumsMap)*2 != len(paths) {
		t.Errorf("unexpected CalculateChecksumsMap() length, got: %d, want: %d", len(checksumsMap), len(paths)/2)
	}

	wantChecksumsMap := make(map[string]string)

	for i := 0; i < len(paths)-1; i = i + 2 {
		wantChecksumsMap[paths[i]] = paths[i+1]
	}

	if !reflect.DeepEqual(checksumsMap, wantChecksumsMap) {
		t.Errorf("unexpected CalculateChecksumsMap() result, got: %v, want: %v", checksumsMap, wantChecksumsMap)
	}
}
