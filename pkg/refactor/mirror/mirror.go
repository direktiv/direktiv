package mirror

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	processStatusComplete  = "complete"
	processStatusPending   = "pending"
	processStatusExecuting = "executing"
	processStatusFailed    = "failed"
)

const (
	processTypeInit = "init"
	processTypeSync = "sync"
)

var ErrNotFound = errors.New("ErrNotFound")

// Config holds configuration data that are needed to create a mirror (pulling mirror credentials, urls, keys
// and any other details).
type Config struct {
	NamespaceID uuid.UUID

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
	ID          uuid.UUID
	NamespaceID uuid.UUID

	Status string
	Typ    string

	EndedAt   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store interface {
	CreateConfig(ctx context.Context, config *Config) (*Config, error)
	UpdateConfig(ctx context.Context, config *Config) (*Config, error)

	GetConfig(ctx context.Context, namespaceID uuid.UUID) (*Config, error)

	CreateProcess(ctx context.Context, process *Process) (*Process, error)
	UpdateProcess(ctx context.Context, process *Process) (*Process, error)

	GetProcess(ctx context.Context, id uuid.UUID) (*Process, error)
	GetProcessesByNamespaceID(ctx context.Context, namespaceID uuid.UUID) ([]*Process, error)

	// TODO: this need to be refactored.
	SetNamespaceVariable(ctx context.Context, namespaceID uuid.UUID, key string, data []byte, hash string, mType string) error

	// TODO: this need to be refactored.
	SetWorkflowVariable(ctx context.Context, workflowID uuid.UUID, key string, data []byte, hash string, mType string) error
}

type Manager interface {
	StartInitialMirroringProcess(ctx context.Context, config *Config) (*Process, error)
	StartSyncingMirrorProcess(ctx context.Context, config *Config) (*Process, error)
	CancelMirroringProcess(ctx context.Context, id uuid.UUID) error
}

type ConfigureWorkflowFunc func(ctx context.Context, file *filestore.File) error

type DefaultManager struct {
	store              Store
	lg                 *zap.SugaredLogger
	fStore             filestore.FileStore
	source             Source
	configWorkflowFunc ConfigureWorkflowFunc
}

func NewDefaultManager(lg *zap.SugaredLogger, store Store, fStore filestore.FileStore, source Source, configWorkflowFunc ConfigureWorkflowFunc) *DefaultManager {
	return &DefaultManager{store: store, lg: lg, fStore: fStore, source: source, configWorkflowFunc: configWorkflowFunc}
}

func (d *DefaultManager) startMirroringProcess(ctx context.Context, config *Config, processType string) (*Process, error) {
	process, err := d.store.CreateProcess(ctx, &Process{
		ID:          uuid.New(),
		NamespaceID: config.NamespaceID,
		Typ:         processType,
		Status:      processStatusPending,
	})
	if err != nil {
		return nil, fmt.Errorf("creating a new process, err: %w", err)
	}

	d.lg.Errorw("starting new mirroring process", "process_id", process.ID, "error", err)

	go func() {
		err := (&mirroringJob{
			ctx: context.TODO(),
			lg:  d.lg,
		}).
			SetProcessStatus(d.store, process, processStatusExecuting).
			CreateTempDirectory().
			PullSourceInPath(d.source, config).
			CreateSourceFilesList().
			// ParseIgnoreFile("/.direktivignore").
			// FilterIgnoredFiles().

			// TODO: we need to implement a mechanism to synchronize multiple mirroring processes.
			ReadRootFilesChecksums(d.fStore, config.NamespaceID).
			CreateAllDirectories(d.fStore, config.NamespaceID).
			CopyFilesToRoot(d.fStore, config.NamespaceID).
			ConfigureWorkflows(d.configWorkflowFunc).
			ParseDirektivVars(d.fStore, d.store, config.NamespaceID).
			CropFilesAndDirectoriesInRoot(d.fStore, config.NamespaceID).
			DeleteTempDirectory().
			SetProcessStatus(d.store, process, processStatusComplete).Error()
		if err != nil {
			process.Status = processStatusFailed
			process, _ = d.store.UpdateProcess(context.TODO(), process)
			d.lg.Errorw("mirroring process failed", "err", err, "process_id", process.ID)

			return
		}

		d.lg.Infow("mirroring process succeeded", "process_id", process.ID)
	}()

	return process, err
}

func (d *DefaultManager) StartInitialMirroringProcess(ctx context.Context, config *Config) (*Process, error) {
	return d.startMirroringProcess(ctx, config, processTypeInit)
}

func (d *DefaultManager) StartSyncingMirrorProcess(ctx context.Context, config *Config) (*Process, error) {
	return d.startMirroringProcess(ctx, config, processTypeSync)
}

//nolint:revive
func (d *DefaultManager) CancelMirroringProcess(ctx context.Context, id uuid.UUID) error {
	// TODO implement me
	return nil
}

var _ Manager = &DefaultManager{}
