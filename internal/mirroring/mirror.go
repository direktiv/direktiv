package mirroring

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/direktiv/direktiv/internal/datastore"
	"github.com/direktiv/direktiv/internal/datastore/datasql"
	"github.com/direktiv/direktiv/internal/telemetry"
	"github.com/direktiv/direktiv/pkg/filestore"
	"github.com/direktiv/direktiv/pkg/filestore/filesql"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-git/go-git/v6/plumbing/format/gitignore"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

const (
	direktivIgnoreFile = ".direktivignore"
)

const (
	keepHours  = 48
	maxRunTime = 2 * time.Minute
)

type mirrorJob struct {
	db *gorm.DB

	process        *datastore.MirrorProcess
	tempDirectory  string
	tempFSRootName string
	matcher        gitignore.Matcher

	err error
}

func RunCleanMirrorProcesses(ctx context.Context, db *gorm.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	maxRecordTime := time.Hour * time.Duration(keepHours)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// clean old mirror processes
			err := datasql.NewStore(db).Mirror().DeleteOldProcesses(ctx, time.Now().Add(-1*maxRecordTime))
			if err != nil {
				slog.Error("could not get mirror processes", slog.Any("error", err))

				continue
			}

			// force unfinished processes
			procs, err := datasql.NewStore(db).Mirror().GetUnfinishedProcesses(ctx)
			if err != nil {
				slog.Error("failed to query unfinished mirror processes", slog.Any("error", err))

				continue
			}

			for _, proc := range procs {
				if time.Since(proc.CreatedAt) > maxRunTime {
					p, err := datasql.NewStore(db).Mirror().GetProcess(ctx, proc.ID)
					if err != nil {
						slog.Error("failed to fetch unfinished mirror process", slog.Any("error", err))

						continue
					}

					if p.Status != datastore.ProcessStatusFailed && p.Status != datastore.ProcessStatusComplete {
						p.Status = datastore.ProcessStatusFailed
						telemetry.LogActivityError(p.Namespace, p.ID.String(), "mirror processing timed out", fmt.Errorf("timed out"))
						_, err = datasql.NewStore(db).Mirror().UpdateProcess(context.Background(), p)
						if err != nil {
							slog.Error("failed to updates status of mirror process", slog.Any("error", err))

							continue
						}
						telemetry.LogNamespaceError(p.Namespace, "mirror processing timed out", fmt.Errorf("timed out"))
					}
				}
			}
		}
	}
}

func MirrorExec(ctx context.Context, db *gorm.DB, cfg *datastore.MirrorConfig, typ string) (*datastore.MirrorProcess, error) {
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
		db:      db,
		process: pro,
	}

	go func() {
		job.setProcessStatus(datastore.ProcessStatusExecuting)
		job.createTempDirectory()
		job.pullSourceIntoTempDirectory(GitSource{}, cfg)
		job.loadGitIgnore()
		job.createTempFSRoot()
		job.copyFilesToTempFSRoot()
		job.deleteTempDirectory()
		job.swapFSRoots()

		job.process.EndedAt = time.Now()
		if job.err == nil {
			job.setProcessStatus(datastore.ProcessStatusComplete)
			return
		}

		telemetry.LogActivity(telemetry.LogLevelError, job.process.Namespace,
			job.process.ID.String(), fmt.Sprintf("mirroring failed '%v'", job.err))
		job.setProcessStatus(datastore.ProcessStatusFailed)
	}()

	return pro, nil
}

func (j *mirrorJob) setProcessStatus(status string) {
	telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace,
		j.process.ID.String(), fmt.Sprintf("mirroring status set to '%s'", status))
	telemetry.LogNamespace(telemetry.LogLevelInfo, j.process.Namespace,
		fmt.Sprintf("mirroring status set to '%s' for %v", status, j.process.ID))

	var err error
	j.process.Status = status
	j.process, err = datasql.NewStore(j.db).Mirror().UpdateProcess(context.Background(), j.process)
	if err != nil {
		j.err = fmt.Errorf("setProcessStatus: datastore set mirror process status: %w", err)
		return
	}
}

func (j *mirrorJob) createTempDirectory() {
	if j.err != nil {
		return
	}

	telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace,
		j.process.ID.String(), "creating temporary directory")

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

func (j *mirrorJob) loadGitIgnore() {
	if j.err != nil {
		return
	}

	j.matcher = gitignore.NewMatcher([]gitignore.Pattern{})

	telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace,
		j.process.ID.String(), "detecting .direktivignore")

	f, err := os.Open(filepath.Join(j.tempDirectory, direktivIgnoreFile))
	if errors.Is(err, os.ErrNotExist) {
		telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace, j.process.ID.String(),
			"no .direktivignore file detected")

		return
	}

	if err != nil {
		j.err = err
		return
	}
	defer f.Close()

	var ps []gitignore.Pattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()
		if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, nil))
		}
	}

	j.matcher = gitignore.NewMatcher(ps)
}

func (j *mirrorJob) pullSourceIntoTempDirectory(source GitSource, cfg *datastore.MirrorConfig) {
	if j.err != nil {
		return
	}

	telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace,
		j.process.ID.String(), fmt.Sprintf("cloning repository with %s", cfg.AuthType))

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

	telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace,
		j.process.ID.String(), "creating temporary fs root")

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
		// skip if in directivignore
		if j.matcher.Match(strings.Split(path, "/"), true) {
			telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace,
				j.process.ID.String(), fmt.Sprintf("direktivignore: skipping path '%s'", path))

			continue
		}

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
		// ignore the ignore file
		if path == "/.direktivignore" {
			continue
		}

		if j.matcher.Match(strings.Split(path, "/"), false) {
			telemetry.LogActivity(telemetry.LogLevelInfo, j.process.Namespace,
				j.process.ID.String(), fmt.Sprintf("direktivignore: skipping path '%s'", path))

			continue
		}

		var data []byte
		data, err = os.ReadFile(j.tempDirectory + "/" + path)
		if err != nil {
			j.err = fmt.Errorf("copyFilesToTempFSRoot: reading os file err: %w", err)
			return
		}

		// default file type
		ft := filestore.FileTypeFile

		var mimeType string
		if strings.HasSuffix(path, "wf.ts") {
			mimeType = "application/x-typescript"
			ft = filestore.FileTypeWorkflow
		} else if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			mimeType = "application/yaml"

			// detect direktiv mimetypes
			ft, err = j.detectDirektivYAML(path, data)
			if err != nil {
				telemetry.LogActivity(telemetry.LogLevelWarn, j.process.Namespace,
					j.process.ID.String(), fmt.Sprintf("detecing yaml failed: %v", err))
			}
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
			ft,
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

const (
	GatewayAPIV1  = "gateway/v1"
	EndpointAPIV2 = "endpoint/v2"

	ServiceAPIV1  = "service/v1"
	ConsumerAPIV1 = "consumer/v1"
	PageAPIV1     = "page/v1"
)

func (j *mirrorJob) detectDirektivYAML(path string, data []byte) (filestore.FileType, error) {
	// detector struct
	type detect struct {
		DirektivAPI  string `yaml:"direktiv_api"`
		XDirektivAPI string `yaml:"x-direktiv-api"`

		// attributes for guessing
		Image string `yaml:"image"`

		// gateway file / gateway/v1
		OpenAPI string `yaml:"openapi"`

		// endpoint
		Methods []string `yaml:"methods"`

		// consumer file
		Username string `yaml:"username"`
		APIKey   string `yaml:"api_key"`
	}

	var a detect
	err := yaml.Unmarshal(data, &a)
	if err != nil {
		return filestore.FileTypeFile, err
	}

	switch a.DirektivAPI {
	case ConsumerAPIV1:
		return filestore.FileTypeConsumer, nil
	case ServiceAPIV1:
		return filestore.FileTypeService, nil
	case PageAPIV1:
		return filestore.FileTypePage, nil
	}

	switch a.XDirektivAPI {
	case GatewayAPIV1:
		return filestore.FileTypeGateway, nil
	case EndpointAPIV2:
		return filestore.FileTypeEndpoint, nil
	}

	// guess file type
	if a.Image != "" {
		telemetry.LogActivity(telemetry.LogLevelWarn, j.process.Namespace,
			j.process.ID.String(), fmt.Sprintf("guessing yaml as direktiv service for file %s", path))

		return filestore.FileTypeService, nil
	}

	if a.Username != "" || a.APIKey != "" {
		telemetry.LogActivity(telemetry.LogLevelWarn, j.process.Namespace,
			j.process.ID.String(), fmt.Sprintf("guessing yaml as direktiv consumer for file %s", path))

		return filestore.FileTypeConsumer, nil
	}

	if len(a.Methods) > 0 {
		telemetry.LogActivity(telemetry.LogLevelWarn, j.process.Namespace,
			j.process.ID.String(), fmt.Sprintf("guessing yaml as direktiv endpoint for file %s", path))

		return filestore.FileTypeEndpoint, nil
	}

	if a.OpenAPI != "" {
		telemetry.LogActivity(telemetry.LogLevelWarn, j.process.Namespace,
			j.process.ID.String(), fmt.Sprintf("guessing yaml as direktiv gateway for file %s", path))

		return filestore.FileTypeGateway, nil
	}

	return filestore.FileTypeFile, nil
}
