package filestore_test

import (
	"context"
	"testing"

	"github.com/direktiv/direktiv/pkg/refactor/database"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/direktiv/direktiv/pkg/refactor/filestore/filestoresql"
	"github.com/google/uuid"
)

func TestRoot_CorrectSetPath(t *testing.T) {
	db, err := database.NewMockGorm()
	if err != nil {
		t.Fatalf("unepxected NewMockGorm() error = %v", err)
	}
	fs := filestoresql.NewSQLFileStore(db)

	tests := []struct {
		name  string
		paths []string

		getPath string
		setPath string

		pathsAfterChange []string
	}{
		{
			name: "basic_directory_case",
			paths: []string{
				"/",
				"/a",
				"/a/b",
				"/a/b/file1.text",
				"/a/c",
				"/a/c/file2.text",
			},

			getPath: "/a/b",
			setPath: "/a/d",

			pathsAfterChange: []string{
				"/",
				"/a",
				"/a/c",
				"/a/c/file2.text",
				"/a/d",
				"/a/d/file1.text",
			},
		},

		{
			name: "weird_case_directory_1", // with /a/bfile1.text
			paths: []string{
				"/",
				"/a",
				"/a/b",
				"/a/b/file1.text",
				"/a/bfile1.text",
				"/a/c",
				"/a/c/file2.text",
			},

			getPath: "/a/b",
			setPath: "/a/d",

			pathsAfterChange: []string{
				"/",
				"/a",
				"/a/bfile1.text",
				"/a/c",
				"/a/c/file2.text",
				"/a/d",
				"/a/d/file1.text",
			},
		},

		{
			name: "weird_case_directory_2", // with /a/b/a/b
			paths: []string{
				"/",
				"/a",
				"/a/b",
				"/a/b/a",
				"/a/b/a/b",
				"/a/b/a/b/file1.text",
				"/a/b/a/c",
				"/a/b/a/c/file2.text",
			},

			getPath: "/a/b/a/b",
			setPath: "/a/b/a/d",

			pathsAfterChange: []string{
				"/",
				"/a",
				"/a/b",
				"/a/b/a",
				"/a/b/a/c",
				"/a/b/a/c/file2.text",
				"/a/b/a/d",
				"/a/b/a/d/file1.text",
			},
		},

		{
			name: "weird_case_directory_3", // with /a/b/a/b
			paths: []string{
				"/",
				"/a",
				"/a/b",
				"/a/b/a",
				"/a/b/a/b",
				"/a/b/a/b/file1.text",
				"/a/b/a/c",
				"/a/b/a/c/file2.text",
			},

			getPath: "/a",
			setPath: "/z",

			pathsAfterChange: []string{
				"/",
				"/z",
				"/z/b",
				"/z/b/a",
				"/z/b/a/b",
				"/z/b/a/b/file1.text",
				"/z/b/a/c",
				"/z/b/a/c/file2.text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := fs.CreateRoot(context.Background(), uuid.New(), uuid.New(), "test")
			if err != nil {
				t.Fatalf("unepxected CreateRoot() error = %v", err)
			}

			for _, path := range tt.paths {
				assertRootCorrectFileCreation(t, fs, root.ID, path)
			}

			assertAllPathsInRoot(t, fs, root.ID, tt.paths...)

			file, err := fs.ForRootID(root.ID).GetFile(context.Background(), tt.getPath)
			if err != nil {
				t.Fatalf("unepxected GetFile() error = %v", err)
			}

			err = fs.ForFile(file).SetPath(context.Background(), tt.setPath)
			if err != nil {
				t.Fatalf("unepxected SetPath() error = %v", err)
			}

			assertAllPathsInRoot(t, fs, root.ID, tt.pathsAfterChange...)
		})
	}
}

func assertAllPathsInRoot(t *testing.T, fs filestore.FileStore, rootID uuid.UUID, wantPaths ...string) {
	t.Helper()

	gotPaths, err := fs.ForRootID(rootID).ListAllFiles(context.Background())
	if err != nil {
		t.Errorf("unexpected ListAllFiles() error = %v", err)

		return
	}
	if len(gotPaths) != len(wantPaths) {
		t.Errorf("unexpected ListAllFiles() length, got: %d, want: %d", len(gotPaths), len(wantPaths))

		return
	}

	for i := range gotPaths {
		if gotPaths[i].Path != wantPaths[i] {
			t.Errorf("unexpected gotPaths[%d] , got: >%s<, want: >%s<", i, gotPaths[i].Path, wantPaths[i])

			return
		}
	}
}
