package mirror

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"go.uber.org/zap"
)

// Settings holds configuration data are needed to create a mirror (pulling mirror credentials, urls, keys
// and any other details).
type Settings struct{}

// Source is an interface that represent a mirror source (git repo is a valid mirror source example).
// In Direktiv, a Mirror is a directory of files that sits somewhere (local or remote) and a user wants to mirror (copy)
// his direktiv namespace files from it.
// Source knows how to access the mirror files (connecting to a remote server in case of git) and copy the files in
// the user's direktiv namespace. Parameter 'settings' is used to configure the sourcing (pulling) mirror process.
// Parameter 'dir' specifies the directory where Source should copy the mirror to.
type Source interface {
	// PullInPath pulls (copies) mirror into local directory specified by 'dir' parameter.
	PullInPath(mirrorSettings Settings, dir string) error
}

// ExecuteMirroringProcess pulls mirror from source, store it in local file system and then push it to direktiv
// filestore.
func ExecuteMirroringProcess(
	ctx context.Context, lg *zap.SugaredLogger,
	direktivRoot filestore.Root,
	source Source, settings Settings,
) error {
	// function starts here:

	distDir, err := os.MkdirTemp("", "direktiv_mirrors")
	if err != nil {
		return fmt.Errorf("create mirror dist_directory, err: %w", err)
	}
	defer func() {
		err := os.RemoveAll(distDir)
		if err != nil {
			lg.Errorf("cleaning mirror dist_directory err: %w", err)
		}
	}()

	err = source.PullInPath(settings, distDir)
	if err != nil {
		return fmt.Errorf("mirror pull, err: %w", err)
	}

	lg.Debugw("mirror fetched", "dist_directory", distDir)

	err = filepath.WalkDir(distDir, func(path string, d fs.DirEntry, err error) error {
		lg = lg.With("path", path, "is_dir", d.IsDir())

		if err != nil {
			return fmt.Errorf("mirror file walk, err: %w", err)
		}

		//nolint
		var fileReader io.ReadCloser = nil
		if !d.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("mirror file walk, read os file, err: %w", err)
			}
			fileReader = io.NopCloser(bytes.NewReader(data))
			defer fileReader.Close()
		}

		// create file(or dir) in directive file store.
		if d.IsDir() {
			_, err = direktivRoot.CreateFile(ctx, strings.TrimPrefix(path, distDir), filestore.FileTypeDirectory, nil)
		} else {
			_, err = direktivRoot.CreateFile(ctx, strings.TrimPrefix(path, distDir), filestore.FileTypeFile, fileReader)
		}

		if err != nil {
			return fmt.Errorf("mirror create filestore entry, err: %w", err)
		}
		lg.Debugw("mirror file saved in filestore")

		return nil
	})

	if err != nil {
		return fmt.Errorf("mirror file walk, err: %w", err)
	}

	lg.Infow("mirror saved successfully",
		"direktiv_root_id", direktivRoot.GetID(), "dist_directory", distDir)

	return nil
}
