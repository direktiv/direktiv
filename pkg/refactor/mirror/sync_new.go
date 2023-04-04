package mirror

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	ignore "github.com/sabhiram/go-gitignore"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

func mirroringProcess() (*Process, error) {
	var store Store
	var fStore filestore.FileStore
	var source Source
	var config *Config

	err := (&mirroringJob{}).
		SetProcessState(store, "started").
		CreateDistDirectory().
		PullSourceInPath(source, config).
		CreateSourceFilesList().
		ParseIgnoreFile("/.direktivignore").
		SkipIignoredFiles().
		ParseDirektivVariable().
		CopyFilesToRoot(fStore).
		CropFilesInRoot(fStore).
		DeleteDistDirectory().
		SetProcessState(store, "finished").Error()

	return nil, err
}

type mirroringJob struct {
	ctx         context.Context
	namespaceID uuid.UUID
	lg          *zap.SugaredLogger
	err         error
	process     *Process
	cancelState *atomic.Uint32

	// JobArtifacts
	distDirectory string
	sourcedPaths  []string
	ignore        *ignore.GitIgnore
}

func (j *mirroringJob) SetProcessState(store Store, state string) *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	j.process.Status = state
	j.process, err = store.UpdateProcess(j.ctx, j.process)

	if err != nil {
		j.err = fmt.Errorf("updating process state, err: %s", err)
	}

	return j
}

func (j *mirroringJob) CreateDistDirectory() *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	j.distDirectory, j.err = os.MkdirTemp("", "direktiv_mirrors")
	if err != nil {
		j.err = fmt.Errorf("create mirror dst directory, err: %s", err)
	}

	return j
}

func (j *mirroringJob) DeleteDistDirectory() *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	err = os.RemoveAll(j.distDirectory)

	if err != nil {
		j.err = fmt.Errorf("os remove dist directory, dir: %s, err: %s", j.distDirectory, err)
	}

	return j
}

func (j *mirroringJob) PullSourceInPath(source Source, config *Config) *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	err = source.PullInPath(config, j.distDirectory)
	if err != nil {
		j.err = fmt.Errorf("pulling source in path, path:%s, err: %s", j.distDirectory, err)
	}

	return j
}

func (j *mirroringJob) CreateSourceFilesList() *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	paths := []string{}

	err = filepath.WalkDir(j.distDirectory, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		j.err = fmt.Errorf("walking dist directory, path: %s, err: %s", j.distDirectory, err)

		return j
	}
	j.sourcedPaths = paths

	return j
}

func (j *mirroringJob) ParseIgnoreFile(ignorePath string) *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	absoulteIgnoreFilePath := filepath.Join(j.distDirectory, ignorePath)

	// first we check if ignore file exists.
	ignoreFileExist := false
	for _, path := range j.sourcedPaths {
		if path == absoulteIgnoreFilePath {
			ignoreFileExist = true
			break
		}
	}
	if !ignoreFileExist {
		// don't parse as there is not ignore file.
		return j
	}

	j.ignore, err = ignore.CompileIgnoreFile(absoulteIgnoreFilePath)
	if err != nil {
		j.err = fmt.Errorf("parsing ignore file, path: %s, err: %s", absoulteIgnoreFilePath, err)
	}

	return j
}

func (j *mirroringJob) SkipIignoredFiles() *mirroringJob {
	if j.err != nil {
		return j
	}

	// Ignore was not parsed because ignore file is not present.
	if j.ignore == nil {
		return j
	}

	skippedList := []string{}
	for _, path := range j.sourcedPaths {
		if !j.ignore.MatchesPath(path) {
			skippedList = append(skippedList, path)
		}
	}
	j.sourcedPaths = skippedList

	return j
}

func (j *mirroringJob) ParseDirektivVariable() *mirroringJob {
	if j.err != nil {
		return j
	}

	return j
}

func (j *mirroringJob) CopyFilesToRoot(fStore filestore.FileStore) *mirroringJob {
	if j.err != nil {
		return j
	}

	for _, path := range j.sourcedPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			j.err = fmt.Errorf("read os file, path: %s, err: %s", path, err)
			return j

		}
		fileReader := bytes.NewReader(data)

		file, err := fStore.ForRootID(j.namespaceID).GetFile(j.ctx, path)

		if err != nil && err != filestore.ErrNotFound {
			j.err = fmt.Errorf("get file from root, path: %s, err: %s", path, err)
			return j
		}

		if err == filestore.ErrNotFound {
			_, _, err = fStore.ForRootID(j.namespaceID).CreateFile(j.ctx, path, filestore.FileTypeFile, fileReader)
			if err != nil {
				j.err = fmt.Errorf("filestore create file, path: %s, err: %s", path, err)

				return j
			}
			continue
		}

		_, err = fStore.ForFile(file).CreateRevision(j.ctx, "", fileReader)
		if err != nil {
			j.err = fmt.Errorf("filestore create revision, path: %s, err: %s", path, err)

			return j
		}
	}

	return j
}

func (j *mirroringJob) CropFilesInRoot(fStore filestore.FileStore) *mirroringJob {
	if j.err != nil {
		return j
	}

	err := fStore.ForRootID(j.namespaceID).BulkRemoveFilesWithExclude(j.ctx, j.sourcedPaths)
	if err != nil {
		j.err = fmt.Errorf("filestore crop to paths, err: %s", err)

		return j
	}

	return j
}

func (j *mirroringJob) CreateAllDirectories(fStore filestore.FileStore) *mirroringJob {
	if j.err != nil {
		return j
	}

	for _, path := range j.sourcedPaths {
		dir := filepath.Dir(path)
		allParentDirs := splitPathToDirectories(dir)
		for _, d := range allParentDirs {
			_, err := fStore.ForRootID(j.namespaceID).GetFile(j.ctx, d)
			if err == nil {
				continue
			}
			if err != filestore.ErrNotFound {
				j.err = fmt.Errorf("filestore get file, path: %s, err: %s", d, err)

				return j
			}
			// directory was not exist, we need to create it.
			_, _, err = fStore.ForRootID(j.namespaceID).CreateFile(j.ctx, d, filestore.FileTypeDirectory, nil)
			if err != nil {
				j.err = fmt.Errorf("filestore create file, path: %s, err: %s", d, err)

				return j
			}
			continue
		}

	}

	return j
}

func (j *mirroringJob) Error() error {
	return j.err
}

func splitPathToDirectories(dir string) []string {
	list := []string{}

	parts := strings.Split(dir, "/")

	for i := range parts {
		list = append(list, "/"+strings.Join(parts[:i], "/"))
	}

	return list
}
