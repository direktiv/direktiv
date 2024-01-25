package filestore_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/google/uuid"
)

func TestRoot_CreateFileWithoutRootDirectory(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	root, err := fs.CreateRoot(context.Background(), uuid.New(), "ns1")
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	assertRootErrorFileCreation(t, fs, root.ID, "/file1.text", filestore.ErrNoParentDirectory)
	assertRootErrorFileCreation(t, fs, root.ID, "/dir1", filestore.ErrNoParentDirectory)

	assertRootCorrectFileCreation(t, fs, root.ID, "/")
	assertRootCorrectFileCreation(t, fs, root.ID, "/file1.text")
	assertRootCorrectFileCreation(t, fs, root.ID, "/dir1")
}

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
		{"/", ""},
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

func TestRootQuery_IsEmptyDirectory(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	root, err := fs.CreateRoot(context.Background(), uuid.New(), "ns1")
	if err != nil {
		t.Fatalf("unepxected CreateRoot() error = %v", err)
	}

	assertEmptyDirectory(t, fs, root.ID, "/", false, filestore.ErrNotFound)
	assertEmptyDirectory(t, fs, root.ID, "/dir1", false, filestore.ErrNotFound)

	assertRootCorrectFileCreation(t, fs, root.ID, "/")
	assertEmptyDirectory(t, fs, root.ID, "/", true, nil)
	assertEmptyDirectory(t, fs, root.ID, "/dir1", false, filestore.ErrNotFound)

	assertRootCorrectFileCreation(t, fs, root.ID, "/file1.text")
	assertRootCorrectFileCreation(t, fs, root.ID, "/file2.text")

	assertRootCorrectFileCreation(t, fs, root.ID, "/dir1")
	assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/file3.text")
	assertRootCorrectFileCreation(t, fs, root.ID, "/dir1/file4.text")

	assertRootCorrectFileCreation(t, fs, root.ID, "/dir2")

	assertEmptyDirectory(t, fs, root.ID, "/", false, nil)
	assertEmptyDirectory(t, fs, root.ID, "/dir1", false, nil)

	assertEmptyDirectory(t, fs, root.ID, "/dir2", true, nil)
}

func assertEmptyDirectory(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, path string, wantEmpty bool, wantErr error) {
	t.Helper()

	gotEmpty, gotErr := fs.ForRootID(rootID).IsEmptyDirectory(context.Background(), path)
	if !errors.Is(gotErr, wantErr) {
		t.Errorf("unexpected IsEmptyDirectory() error, got: %v, want: %v", gotErr, wantErr)

		return
	}
	if gotEmpty != wantEmpty {
		t.Errorf("unexpected IsEmptyDirectory(), got: %v, want %v", gotEmpty, wantEmpty)
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
		assertRootCorrectFileCreation(t, fs, root.ID, "/")
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
		assertRootCorrectFileCreation(t, fs, root.ID, "/")
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
