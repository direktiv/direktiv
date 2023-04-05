package mirror

import (
	"context"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

// Config holds configuration data that are needed to create a mirror (pulling mirror credentials, urls, keys
// and any other details).
type Config struct {
	ID uuid.UUID

	URL                  string
	GitRef               string
	GitCommitHash        string
	PublicKey            string
	PrivateKey           string
	PrivateKeyPassphrase string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Process struct {
	ID       uuid.UUID
	ConfigID uuid.UUID

	Status string

	EndedAt   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store interface {
	CreateConfig(ctx context.Context, config *Config) (*Config, error)
	UpdateConfig(ctx context.Context, config *Config) (*Config, error)

	GetConfig(ctx context.Context, id uuid.UUID) (*Config, error)

	CreateProcess(ctx context.Context, process *Process) (*Process, error)
	UpdateProcess(ctx context.Context, process *Process) (*Process, error)

	GetProcess(ctx context.Context, id uuid.UUID) (*Process, error)
	GetProcessesByConfig(ctx context.Context, configID uuid.UUID) ([]*Process, error)
}

type Manager interface {
	StartMirroringProcess(ctx context.Context, config *Config) (*Process, error)
	CancelMirroringProcess(ctx context.Context, id uuid.UUID) error
}

type defaultManager struct {
	store Store
	lg    *zap.SugaredLogger
}

func (d *defaultManager) StartMirroringProcess(ctx context.Context, config *Config) (*Process, error) {
	var fStore filestore.FileStore
	var source Source
	var process *Process
	var namespaceID uuid.UUID

	process, err := d.store.CreateProcess(ctx, &Process{
		ConfigID: config.ID,
		Status:   "created",
	})
	if err != nil {
		return nil, fmt.Errorf("creating a new process, err: %s", err)
	}

	d.lg.Errorw("starting new mirroring process", "process_id", process.ID)

	go func() {
		err := (&mirroringJob{
			ctx: context.TODO(),
			lg:  d.lg,
		}).
			SetProcessStatus(d.store, process, "started").
			CreateDistDirectory().
			PullSourceInPath(source, config).
			CreateSourceFilesList().
			ParseIgnoreFile("/.direktivignore").
			FilterIgnoredFiles().
			ParseDirektivVariable().

			// TODO: we need to implement a mechanism to synchronize multiple mirroring processes.
			ReadRootFilesChecksums(fStore, namespaceID).
			CreateAllDirectories(fStore, namespaceID).
			CopyFilesToRoot(fStore, namespaceID).
			CropFilesAndDirectoriesInRoot(fStore, namespaceID).
			DeleteDistDirectory().
			SetProcessStatus(d.store, process, "finished").Error()
		if err != nil {
			process.Status = "failed"
			process, err = d.store.UpdateProcess(context.TODO(), process)
		}
		if err != nil {
			d.lg.Errorw("mirroring process failed", "err", err, "process_id", process.ID)
		}
		if err == nil {
			d.lg.Errorw("mirroring process succeeded", "process_id", process.ID)
		}
	}()

	return process, err
}

func (d *defaultManager) CancelMirroringProcess(ctx context.Context, id uuid.UUID) error {
	// TODO implement me
	return nil
}

var _ Manager = &defaultManager{}
