package mirror

import (
	"bytes"
	"context"
	"fmt"
	"github.com/direktiv/direktiv/pkg/refactor/ignorefile"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"go.uber.org/zap"
)

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
	fStore filestore.FileStore,
	direktivRoot *filestore.Root,
	ignoreMatcher ignorefile.Matcher,
	source Source, settings Settings,
) error {
	// function starts here:
	dstDir, err := os.MkdirTemp("", "direktiv_mirrors")
	if err != nil {
		return fmt.Errorf("create mirror dst_directory, err: %w", err)
	}
	defer func() {
		err := os.RemoveAll(dstDir)
		if err != nil {
			lg.Errorf("cleaning mirror dist_directory err: %w", err)
		}
	}()

	err = source.PullInPath(settings, dstDir)
	if err != nil {
		return fmt.Errorf("mirror pull, err: %w", err)
	}

	lg.Debugw("mirror fetched", "dist_directory", dstDir)

	return ExecuteMirroringProcessFromDirectory(ctx, lg, fStore, direktivRoot, ignoreMatcher, dstDir)
}

func ExecuteMirroringProcessFromDirectory(
	ctx context.Context, lg *zap.SugaredLogger,
	fStore filestore.FileStore,
	direktivRoot *filestore.Root,
	ignoreMatcher ignorefile.Matcher,
	dstDir string,
) error {
	currentChecksumsMap, err := fStore.ForRoot(direktivRoot).CalculateChecksumsMap(ctx, "/")
	if err != nil {
		return fmt.Errorf("calculate current checksum map, err: %w", err)
	}

	err = filepath.WalkDir(dstDir, func(path string, d fs.DirEntry, err error) error {
		lg = lg.With("path", path, "is_dir", d.IsDir())

		if err != nil {
			return fmt.Errorf("mirror file walk, err: %w", err)
		}

		relativePath := strings.TrimPrefix(path, dstDir)

		if ignoreMatcher.MatchPath(relativePath) {
			return nil
		}

		defer delete(currentChecksumsMap, relativePath)

		var fileReader io.Reader
		var checksum string
		if !d.IsDir() {
			var err error
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("mirror file walk, open os file, err: %w", err)
			}
			checksum = string(filestore.DefaultCalculateChecksum(data))
			fileReader = bytes.NewReader(data)
		}

		fileChecksum, doesPathExists := currentChecksumsMap[relativePath]
		isDifferentChecksum := checksum != fileChecksum

		// create dir in directive file store.
		if d.IsDir() && !doesPathExists {
			_, err = fStore.
				ForRoot(direktivRoot).
				CreateFile(ctx, relativePath, filestore.FileTypeDirectory, nil)
		}
		if !d.IsDir() && isDifferentChecksum {
			_, err = fStore.
				ForRoot(direktivRoot).
				CreateFile(ctx, relativePath, filestore.FileTypeFile, fileReader)
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
		"direktiv_root_id", direktivRoot.ID, "dist_directory", dstDir)

	return nil
}
