package filestore_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/database"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filestoresql"
	"github.com/google/uuid"
)

func TestRoot_CreateFile(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	root, err := fs.CreateRoot(context.Background(), uuid.New(), "ns1")
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	tests := []struct {
		path    string
		payload string
	}{
		{"/example.text", "abcd"},
		{"/example1.text", "abcd"},
		{"/example2.text", "abcd"},
	}
	for _, tt := range tests {
		t.Run("valid", func(t *testing.T) {
			assertRootCorrectFileCreation(t, fs, root.ID, tt.path)
		})
	}
}

func assertRootCorrectFileCreation(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, path string) {
	t.Helper()

	var data []byte
	typ := filestore.FileTypeDirectory
	if strings.Contains(path, ".text") {
		data = []byte("some data")
		typ = filestore.FileTypeFile
	}

	assertRootCorrectFileCreationWithContent(t, fs, rootID, path, string(typ), data)
}

func assertRootCorrectFileCreationWithContent(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, path string, typ string, data []byte) {
	t.Helper()

	file, err := fs.ForRootID(rootID).CreateFile(context.Background(), path, filestore.FileType(typ), "application/octet-stream", data)
	if err != nil {
		t.Errorf("unexpected CreateFile() error: %v", err)

		return
	}
	if file == nil {
		t.Errorf("unexpected nil file CreateFile()")

		return
	}
	if file.Path != path {
		t.Errorf("unexpected file.Path, got: >%s<, want: >%s<", file.Path, path)

		return
	}

	if typ != "directory" {
		createdData, _ := fs.ForFile(file).GetData(context.Background())
		if string(createdData) != string(data) {
			t.Errorf("unexpected GetPath(), got: >%s<, want: >%s<", createdData, data)

			return
		}
	}

	file, err = fs.ForRootID(rootID).GetFile(context.Background(), path)
	if err != nil {
		t.Errorf("unexpected GetFile() error: %v", err)

		return
	}
	if file == nil {
		t.Errorf("unexpected nil file GetFile()")

		return
	}
	if file.Path != path {
		t.Errorf("unexpected file.Path, got: >%s<, want: >%s<", file.Path, path)

		return
	}
}

func assertRootErrorFileCreation(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, path string, wantErr error) {
	t.Helper()

	typ := filestore.FileTypeDirectory
	if strings.Contains(path, ".text") {
		typ = filestore.FileTypeFile
	}

	file, gotErr := fs.ForRootID(rootID).CreateFile(context.Background(), path, typ, "application/octet-stream", []byte(""))
	if file != nil {
		t.Errorf("unexpected none nil CreateFile().file")

		return
	}
	if !errors.Is(gotErr, wantErr) {
		t.Errorf("unexpected CreateFile() error, got: %v, want: %v", gotErr, wantErr)

		return
	}
}

func TestRoot_CorrectReadDirectory(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	root, err := fs.CreateRoot(context.Background(), uuid.New(), "ns1")
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	// Test root directory:
	{
		assertRootCorrectFileCreation(t, fs, root.ID, "/file1.text")
		assertRootCorrectFileCreation(t, fs, root.ID, "/file2.text")

		assertRootFilesInPath(t, fs, root.ID, "/",
			"/file1.text",
			"/file2.text",
		)
	}

	// Add /dir1 directory:
	{
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1")
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/file3.text")
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/file4.text")

		assertRootFilesInPath(t, fs, root.ID, "/dir1",
			"/dir1/file3.text",
			"/dir1/file4.text",
		)
		assertRootFilesInPath(t, fs, root.ID, "/",
			"/dir1",
			"/file1.text",
			"/file2.text",
		)
	}

	// Add /dir1/dir2 directory:
	{
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/dir2")
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/dir2/file5.text")
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/dir2/file6.text")

		assertRootFilesInPath(t, fs, root.ID, "/dir1/dir2",
			"/dir1/dir2/file5.text",
			"/dir1/dir2/file6.text",
		)
		assertRootFilesInPath(t, fs, root.ID, "/dir1",
			"/dir1/dir2",
			"/dir1/file3.text",
			"/dir1/file4.text",
		)
		assertRootFilesInPath(t, fs, root.ID, "/",
			"/dir1",
			"/file1.text",
			"/file2.text",
		)
	}
}

func TestRoot_RenamePath(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	root, err := fs.CreateRoot(context.Background(), uuid.New(), "ns1")
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	// Test root directory:
	{
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1")
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/dir2")
		assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/file.text")
	}

	f, err := fs.ForRootID(root.ID).GetFile(context.Background(), "/dir1/file.text")
	if err != nil {
		t.Fatalf("unepxected GetFile() error = %v", err)
	}
	err = fs.ForFile(f).SetPath(context.Background(), "/file.text")
	if err != nil {
		t.Fatalf("unepxected SetPath() error = %v", err)
	}

	assertRootFilesInPath(t, fs, root.ID, "/",
		"/dir1",
		"/file.text",
	)

	f, err = fs.ForRootID(root.ID).GetFile(context.Background(), "/dir1/dir2")
	if err != nil {
		t.Fatalf("unepxected GetFile() error = %v", err)
	}
	err = fs.ForFile(f).SetPath(context.Background(), "/dir2")
	if err != nil {
		t.Fatalf("unepxected SetPath() error = %v", err)
	}

	assertRootFilesInPath(t, fs, root.ID, "/",
		"/dir1",
		"/dir2",
		"/file.text",
	)
}

func assertRootFilesInPath(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, searchPath string, paths ...string) {
	t.Helper()

	files, err := fs.ForRootID(rootID).ReadDirectory(context.Background(), searchPath)
	if err != nil {
		t.Errorf("unepxected ReadDirectory() error = %v", err)

		return
	}
	if len(files) != len(paths) {
		t.Errorf("unexpected ReadDirectory() length, got: %d, want: %d", len(files), len(paths))

		return
	}

	for i := range paths {
		if files[i].Path != paths[i] {
			t.Errorf("unexpected files[%d].Path , got: >%s<, want: >%s<", i, files[i].Path, paths[i])

			return
		}
	}
}

func assertFileExistsV2(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, file filestore.File) {
	t.Helper()

	f, err := fs.ForRootID(rootID).GetFile(context.Background(), file.Path)
	if err != nil {
		t.Errorf("unexpected GetFile() error: %v", err)

		return
	}
	if f == nil {
		t.Errorf("unexpected nil file GetFile()")

		return
	}
	if f.Path != file.Path {
		t.Errorf("unexpected file.Path, got: >%s<, want: >%s<", f.Path, file.Path)

		return
	}
	if f.Typ != file.Typ {
		t.Errorf("unexpected file.Typ, got: >%s<, want: >%s<", f.Typ, file.Typ)

		return
	}
	if f.Typ == filestore.FileTypeDirectory {
		return
	}
	data, err := fs.ForFile(f).GetData(context.Background())
	if err != nil {
		t.Errorf("unexpected GetData() error: %v", err)

		return
	}
	if data == nil {
		t.Errorf("unexpected nil data GetData()")

		return
	}
	if string(data) != string(file.Data) {
		t.Errorf("unexpected data, got: >%s<, want: >%s<", string(data), string(file.Data))

		return
	}
}

func assertCreateFileV2(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, file filestore.File) {
	t.Helper()

	f, err := fs.ForRootID(rootID).CreateFile(context.Background(), file.Path, file.Typ, "text/plain", file.Data)
	if err != nil {
		t.Errorf("unexpected CreateFile() error: %v", err)

		return
	}
	if f == nil {
		t.Errorf("unexpected nil file CreateFile()")

		return
	}
}
