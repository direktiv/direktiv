package mirroring

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

type mirrorJob struct {
	manager *Manager
	db      *gorm.DB

	process        *datastore.MirrorProcess
	tempDirectory  string
	tempFSRootName string

	err error
}

func (m *Manager) Exec(ctx context.Context, db *gorm.DB, cfg *datastore.MirrorConfig, typ string) (*datastore.MirrorProcess, error) {
	pro := &datastore.MirrorProcess{
		ID:        uuid.New(),
		Namespace: cfg.Namespace,
		Typ:       typ,
		Status:    datastore.ProcessStatusPending,
	}

	var err error
	pro, err = datasql.NewStore(db).Mirror().CreateProcess(ctx, pro)
	if err != nil {
		return nil, fmt.Errorf("datastore create mirror process: %w", err)
	}

	job := &mirrorJob{
		manager: m,
		db:      db,
		process: pro,
	}

	go func() {
		job.setProcessStatue(datastore.ProcessStatusExecuting)
		job.createTempDirectory()
		job.pullSourceIntoTempDirectory(GitSource{}, cfg)
		job.createTempFSRoot()
		job.copyFilesToTempFSRoot()
		job.deleteTempDirectory()
		job.swapFSRoots()
		job.setProcessStatue(datastore.ProcessStatusComplete)

		if job.err == nil {
			return
		}

		job.setProcessStatue(datastore.ProcessStatusFailed)

		err = job.err
		if err != nil {
			fmt.Printf(">>>>> error creating mirror process: %v\n", err)
		}
	}()

	return pro, nil
}

func (j *mirrorJob) setProcessStatue(status string) {
	if j.err != nil {
		return
	}

	var err error
	j.process.Status = status
	j.process, err = datasql.NewStore(j.db).Mirror().UpdateProcess(context.Background(), j.process)

	if err != nil {
		j.err = fmt.Errorf("setProcessStatue: datastore set mirror process status: %w", err)
		return
	}
}

func (j *mirrorJob) createTempDirectory() {
	if j.err != nil {
		return
	}

	var err error
	j.tempDirectory, err = os.MkdirTemp("", "direktiv_mirrors")
	if err != nil {
		j.err = fmt.Errorf("createTempDirectory: create tempDirectory directory, err: %w", err)
		return
	}
}

func (j *mirrorJob) deleteTempDirectory() {
	if j.err != nil {
		return
	}

	err := os.RemoveAll(j.tempDirectory)
	if err != nil {
		j.err = fmt.Errorf("deleteTempDirectory: os remove tempDirectory directory, dir: %s, err: %w", j.tempDirectory, err)
		return
	}
}

func (j *mirrorJob) pullSourceIntoTempDirectory(source GitSource, cfg *datastore.MirrorConfig) {
	if j.err != nil {
		return
	}

	err := source.PullInPath(cfg, j.tempDirectory)
	if err != nil {
		j.err = fmt.Errorf("pullSourceIntoTempDirectory: mirroring source pull: %w", err)
		return
	}
}

func (j *mirrorJob) createTempFSRoot() {
	if j.err != nil {
		return
	}

	j.tempFSRootName = uuid.New().String()
	_, err := filesql.NewStore(j.db).CreateRoot(context.Background(), j.tempFSRootName)
	if err != nil {
		j.err = fmt.Errorf("createTempFSRoot: creating fs root, err: %w", err)
		return
	}
}

func (j *mirrorJob) copyFilesToTempFSRoot() {
	if j.err != nil {
		return
	}

	var createDirs []string
	var createsFiles []string

	err := filepath.WalkDir(j.tempDirectory, func(path string, d os.DirEntry, err error) error {
		path = strings.TrimPrefix(path, j.tempDirectory)

		if path == "" || path == "/" || path == "." || strings.HasPrefix(path, "/.git") {
			return nil
		}

		if d.IsDir() {
			createDirs = append(createDirs, path)
		} else {
			createsFiles = append(createsFiles, path)
		}

		return nil
	})

	if err != nil {
		j.err = fmt.Errorf("copyFilesToTempFSRoot: walk dir err: %w", err)
		return
	}

	for _, path := range createDirs {
		fmt.Printf(">>>>> ddd>%s<\n", path)
	}
	for _, path := range createsFiles {
		fmt.Printf(">>>>> fff>%s<\n", path)
	}

	for _, path := range createDirs {
		_, err = filesql.NewStore(j.db).ForRoot(j.tempFSRootName).CreateFile(
			context.Background(),
			path,
			filestore.FileTypeDirectory,
			"",
			nil)
		if err != nil {
			j.err = fmt.Errorf("copyFilesToTempFSRoot: creating fs dir err: %w", err)
			return
		}
	}

	for _, path := range createsFiles {
		var data []byte
		data, err = os.ReadFile(j.tempDirectory + "/" + path)
		if err != nil {
			j.err = fmt.Errorf("copyFilesToTempFSRoot: reading os file err: %w", err)
			return
		}

		var mimeType string
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			mimeType = "application/yaml"
		} else {
			mt := mimetype.Detect(data)
			mimeType = strings.Split(mt.String(), ";")[0]

		}
		if mimeType == "" {
			mimeType = "application/octet-stream" // fallback
		}

		_, err = filesql.NewStore(j.db).ForRoot(j.tempFSRootName).CreateFile(
			context.Background(),
			path,
			filestore.FileTypeFile,
			mimeType,
			data)
		if err != nil {
			j.err = fmt.Errorf("copyFilesToTempFSRoot: creating fs file err: %w", err)
			return
		}
	}
}

func (j *mirrorJob) swapFSRoots() {
	if j.err != nil {
		return
	}

	db := j.db.Begin()
	if db.Error != nil {
		j.err = fmt.Errorf("swapFSRoots: begin dbtx: %w", db.Error)
	}
	defer db.Rollback()

	fs := filesql.NewStore(db)

	err := fs.ForRoot(j.process.Namespace).Delete(context.Background())
	if err != nil {
		j.err = fmt.Errorf("swapFSRoots: deleting fs root err: %w", err)
		return
	}

	err = fs.ForRoot(j.tempFSRootName).SetID(context.Background(), j.process.Namespace)
	if err != nil {
		j.err = fmt.Errorf("swapFSRoots: setting fs root err: %w", err)
		return
	}

	err = db.Commit().Error
	if err != nil {
		j.err = fmt.Errorf("swapFSRoots: commit dbtx: %w", err)
		return
	}
}
