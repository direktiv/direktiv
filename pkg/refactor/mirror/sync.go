package mirror

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	ignore "github.com/sabhiram/go-gitignore"
)

// TODO: implement parsing direktiv variables.
// TODO: check %w verb on errors.
// TODO: fix errors and add logs.
// TODO: implement a mechanism to clean dangling processes and cleaning them up.
// TODO: implement synchronizing jobs.

// mirroringJob implements a unique pattern. mirroringJob is a struct with both input fields and artifact fields.
// various methods get called on *mirroringJob. Each method deliver its functionality by mutate artifact fields of
// *mirroringJob. Errors of methods of *mirroringJob will not be returned but will be set in *mirroringJob.error.
type mirroringJob struct {
	// job parameters.

	//nolint:containedctx
	ctx         context.Context
	infoLogFunc LogFunc

	// job artifacts.
	processID             uuid.UUID
	err                   error
	distDirectory         string
	sourcedPaths          []string
	direktivIgnore        *ignore.GitIgnore
	rootChecksums         map[string]string
	changedOrNewWorkflows []*filestore.File
}

// SetProcessID sets mirroring process ID.
func (j *mirroringJob) SetProcessID(processID uuid.UUID) *mirroringJob {
	if j.err != nil {
		return j
	}
	j.processID = processID

	return j
}

// SetProcessStatus sets mirroring process status.
func (j *mirroringJob) SetProcessStatus(store Store, process *Process, status string) *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	process.Status = status
	if status == processStatusComplete || status == processStatusFailed {
		process.EndedAt = time.Now()
	}
	_, err = store.UpdateProcess(j.ctx, process)

	if err != nil {
		j.err = fmt.Errorf("updating process state, err: %w", err)
	}

	return j
}

// CreateTempDirectory creates a new os temp directory so that the mirror files with be sourced in.
func (j *mirroringJob) CreateTempDirectory() *mirroringJob {
	if j.err != nil {
		return j
	}
	var err error

	j.distDirectory, j.err = os.MkdirTemp("", "direktiv_mirrors")
	if err != nil {
		j.err = fmt.Errorf("create mirror dst directory, err: %w", err)
	}

	j.logInfo("creating mirroring temp dir", "dir", j.distDirectory)

	return j
}

// DeleteTempDirectory cleanup the os temp directory that the mirror files was sourced in.
func (j *mirroringJob) DeleteTempDirectory() *mirroringJob {
	if j.err != nil {
		return j
	}

	err := os.RemoveAll(j.distDirectory)
	if err != nil {
		j.err = fmt.Errorf("os remove dist directory, dir: %s, err: %w", j.distDirectory, err)
	}

	j.logInfo("deleting mirroring temp dir", "dir", j.distDirectory)

	return j
}

// PullSourceInPath pulls the mirror files from the source to the temp os directory.
func (j *mirroringJob) PullSourceInPath(source Source, config *Config) *mirroringJob {
	if j.err != nil {
		return j
	}

	err := source.PullInPath(config, j.distDirectory)
	if err != nil {
		j.err = fmt.Errorf("pulling source in path, path:%s, err: %w", j.distDirectory, err)
	}

	return j
}

// CreateSourceFilesList creates a list of all relevant mirror file paths. The produced list is necessary for
// further mirroring steps.
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

		relativePath := strings.TrimPrefix(path, j.distDirectory)

		if _, err := filestore.SanitizePath(relativePath); err != nil {
			return nil
		}
		paths = append(paths, relativePath)

		return nil
	})
	if err != nil {
		j.err = fmt.Errorf("walking dist directory, path: %s, err: %w", j.distDirectory, err)

		return j
	}
	j.sourcedPaths = paths

	for _, p := range j.sourcedPaths {
		j.logInfo("source path", "path", p)
	}

	return j
}

// ParseIgnoreFile parses the direktiv ignore file if exists.
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

	j.direktivIgnore, err = ignore.CompileIgnoreFile(absoulteIgnoreFilePath)
	if err != nil {
		j.err = fmt.Errorf("parsing ignore file, path: %s, err: %w", absoulteIgnoreFilePath, err)
	}

	return j
}

// FilterIgnoredFiles filters the direktiv ignored files.
func (j *mirroringJob) FilterIgnoredFiles() *mirroringJob {
	if j.err != nil {
		return j
	}

	// Ignore was not parsed because ignore file is not present.
	if j.direktivIgnore == nil {
		return j
	}

	skippedList := []string{}
	for _, path := range j.sourcedPaths {
		if !j.direktivIgnore.MatchesPath(path) {
			skippedList = append(skippedList, path)
		}
	}
	j.sourcedPaths = skippedList

	return j
}

// ParseDirektivVars tries to parse special direktiv files naming convention to create both namespace and workflow
// files.
func (j *mirroringJob) ParseDirektivVars(fStore filestore.FileStore, vStore core.RuntimeVariablesStore, namespaceID, rootID uuid.UUID) *mirroringJob {
	if j.err != nil {
		return j
	}

	namespaceVarKeys, workflowVarKeys := parseDirektivVars(j.sourcedPaths)

	for _, pk := range namespaceVarKeys {
		path := j.distDirectory + pk[0] + "." + pk[1]
		data, err := os.ReadFile(path)
		if err != nil {
			j.err = fmt.Errorf("read os file, path: %s, err: %w", path, err)

			return j
		}
		mType := mimetype.Detect(data)
		if err != nil {
			j.err = fmt.Errorf("calculate hash string, path: %s, err: %w", path, err)

			return j
		}

		_, err = vStore.Set(j.ctx,
			&core.RuntimeVariable{
				NamespaceID: namespaceID,
				Name:        pk[1],
				MimeType:    mType.String(),
				Data:        data,
			})
		if err != nil {
			j.err = fmt.Errorf("save namespace variable, path: %s, err: %w", path, err)

			return j
		}
	}

	for _, pk := range workflowVarKeys {
		path := j.distDirectory + pk[0] + "." + pk[1]
		workflowFile, err := fStore.ForRootID(rootID).GetFile(j.ctx, pk[0])
		if errors.Is(err, filestore.ErrNotFound) {
			continue
		}
		if err != nil {
			j.err = fmt.Errorf("read filestore file, path: %s, err: %w", path, err)

			return j
		}

		data, err := os.ReadFile(path)
		if err != nil {
			j.err = fmt.Errorf("read os file, path: %s, err: %w", path, err)

			return j
		}
		mType := mimetype.Detect(data)
		if err != nil {
			j.err = fmt.Errorf("calculate hash string, path: %s, err: %w", path, err)

			return j
		}
		_, err = vStore.Set(j.ctx,
			&core.RuntimeVariable{
				NamespaceID:  namespaceID,
				WorkflowPath: workflowFile.Path,
				Name:         pk[1],
				MimeType:     mType.String(),
				Data:         data,
			})
		if err != nil {
			j.err = fmt.Errorf("save workflow variable, path: %s, err: %w", path, err)

			return j
		}
	}

	return j
}

// CopyFilesToRoot copies files to the filestore.
func (j *mirroringJob) CopyFilesToRoot(fStore filestore.FileStore, rootID uuid.UUID) *mirroringJob {
	if j.err != nil {
		return j
	}

	for _, path := range j.sourcedPaths {
		j.logInfo("trying to copy", "path", path)
		data, err := os.ReadFile(j.distDirectory + path)
		if err != nil {
			j.err = fmt.Errorf("read os file, path: %s, err: %w", path, err)

			return j
		}
		checksum := string(filestore.DefaultCalculateChecksum(data))
		fileChecksum, pathDoesExist := j.rootChecksums[path]
		isEqualChecksum := checksum == fileChecksum

		if pathDoesExist && isEqualChecksum {
			j.logInfo("checksum skipped to root", "path", path)

			continue
		}

		fileReader := bytes.NewReader(data)

		if !pathDoesExist {
			typ := filestore.FileTypeFile
			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				typ = filestore.FileTypeWorkflow
			}
			file, _, err := fStore.ForRootID(rootID).CreateFile(j.ctx, path, typ, fileReader)
			if err != nil {
				j.err = fmt.Errorf("filestore create file, path: %s, err: %w", path, err)

				return j
			}
			j.logInfo("copied to root", "path", path)

			if file.Typ == filestore.FileTypeWorkflow {
				j.changedOrNewWorkflows = append(j.changedOrNewWorkflows, file)
			}

			continue
		}

		file, err := fStore.ForRootID(rootID).GetFile(j.ctx, path)
		if err != nil {
			j.err = fmt.Errorf("get file from root, path: %s, err: %w", path, err)

			return j
		}

		_, err = fStore.ForFile(file).CreateRevision(j.ctx, "", fileReader)
		if err != nil {
			j.err = fmt.Errorf("filestore create revision, path: %s, err: %w", path, err)

			return j
		}
		j.logInfo("revision to root", "path", path)

		if file.Typ == filestore.FileTypeWorkflow {
			j.changedOrNewWorkflows = append(j.changedOrNewWorkflows, file)
		}
	}

	return j
}

// ConfigureWorkflows calls a function hook for every changed or new workflow.
func (j *mirroringJob) ConfigureWorkflows(nsID uuid.UUID, configureFunc ConfigureWorkflowFunc) *mirroringJob {
	if j.err != nil {
		return j
	}
	if configureFunc == nil {
		return j
	}

	for _, file := range j.changedOrNewWorkflows {
		err := configureFunc(j.ctx, nsID, file)
		if err != nil {
			j.err = fmt.Errorf("configure workflow, path: %s, err: %w", file.Path, err)

			return j
		}
		j.logInfo("workflow configured correctly", "path", file.Path)
	}

	return j
}

// CropFilesAndDirectoriesInRoot crops the filestore to remove all files and directories that don't exist in the mirror.
func (j *mirroringJob) CropFilesAndDirectoriesInRoot(fStore filestore.FileStore, rootID uuid.UUID) *mirroringJob {
	if j.err != nil {
		return j
	}

	err := fStore.ForRootID(rootID).CropFilesAndDirectories(j.ctx, j.sourcedPaths)
	if err != nil {
		j.err = fmt.Errorf("filestore crop to paths, err: %w", err)

		return j
	}

	return j
}

// ReadRootFilesChecksums reads the rootChecksums param which helps to prevent copying none-changed files
// to the filestore.
func (j *mirroringJob) ReadRootFilesChecksums(fStore filestore.FileStore, rootID uuid.UUID) *mirroringJob {
	if j.err != nil {
		return j
	}

	checksums, err := fStore.ForRootID(rootID).CalculateChecksumsMap(j.ctx)
	if err != nil {
		j.err = fmt.Errorf("filestore calculate checksums map, err: %w", err)

		return j
	}

	j.rootChecksums = checksums

	return j
}

// CreateAllDirectories creates all the directories that appears in the mirror.
func (j *mirroringJob) CreateAllDirectories(fStore filestore.FileStore, rootID uuid.UUID) *mirroringJob {
	if j.err != nil {
		return j
	}

	createdDirs := map[string]bool{}

	for _, path := range j.sourcedPaths {
		dir := filepath.Dir(path)
		allParentDirs := splitPathToDirectories(dir)
		for _, d := range allParentDirs {
			if _, isExists := j.rootChecksums[d]; isExists {
				continue
			}

			if _, isCreated := createdDirs[d]; isCreated {
				continue
			}

			_, _, err := fStore.ForRootID(rootID).CreateFile(j.ctx, d, filestore.FileTypeDirectory, nil)

			// check if it is a fatal error.
			if err != nil && !errors.Is(err, filestore.ErrPathAlreadyExists) {
				j.err = fmt.Errorf("filestore create dir, path: %s, err: %w", d, err)

				return j
			}

			createdDirs[d] = true
		}
	}

	return j
}

func (j *mirroringJob) Error() interface{} {
	return j.err
}

func (j *mirroringJob) logInfo(msg string, keysAndValues ...interface{}) {
	j.infoLogFunc(j.processID, msg, keysAndValues...)
}

func splitPathToDirectories(dir string) []string {
	list := []string{}

	dir = strings.TrimSpace(dir)
	dir = strings.TrimPrefix(dir, "/")

	parts := strings.Split(dir, "/")

	for i := range parts {
		list = append(list, "/"+strings.Join(parts[:i+1], "/"))
	}

	return list
}

func parseDirektivVars(paths []string) ([][]string, [][]string) {
	pathsMap := map[string]bool{}
	for _, p := range paths {
		pathsMap[p] = true
	}

	namespaceVarPathsKeys := [][]string{}
	workflowVarPathsKeys := [][]string{}

	for _, p := range paths {
		base := filepath.Base(p)
		dir := filepath.Dir(p)

		if strings.Contains(base, "var.") && len(base) > len("var.") {
			if strings.HasPrefix(strings.TrimPrefix(base, "var."), "_") {
				continue
			}
			namespaceVarPathsKeys = append(namespaceVarPathsKeys, []string{filepath.Clean(dir + "/var"), strings.TrimPrefix(base, "var.")})

			continue
		}

		if strings.Contains(base, ".yaml.") {
			if _, ok := pathsMap[p]; !ok {
				continue
			}
			parts := strings.Split(base, ".yaml.")
			//nolint:gomnd
			if len(parts) == 2 {
				workflowVarPathsKeys = append(workflowVarPathsKeys, []string{dir + "/" + parts[0] + ".yaml", parts[1]})
			}
		}
		if strings.Contains(base, ".yml.") {
			if _, ok := pathsMap[p]; !ok {
				continue
			}
			parts := strings.Split(base, ".yml.")
			//nolint:gomnd
			if len(parts) == 2 {
				workflowVarPathsKeys = append(workflowVarPathsKeys, []string{dir + "/" + parts[0] + ".yml", parts[1]})
			}
		}
	}

	return namespaceVarPathsKeys, workflowVarPathsKeys
}
