package mirror

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/direktiv/direktiv/pkg/refactor/core"
	"github.com/direktiv/direktiv/pkg/refactor/filestore"
	"github.com/google/uuid"
)

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

// Process different statuses.
const (
	processStatusComplete  = "complete"
	processStatusPending   = "pending"
	processStatusExecuting = "executing"
	processStatusFailed    = "failed"
)

// Process different types.
const (
	// Indicates initial mirroring process.
	processTypeInit = "init"

	// Indicates re-mirroring process.
	processTypeSync = "sync"
)

// Process represents an instance of mirroring process that happened or is currently happened. For every mirroring
// process gets executing, a Process instance should be created with mirror.Store.
type Process struct {
	ID          uuid.UUID
	NamespaceID uuid.UUID

	Status string
	Typ    string

	EndedAt   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

var ErrNotFound = errors.New("ErrNotFound")

// Store *doesn't* lunch any mirroring process. Store is only responsible for fetching and setting mirror.Config and
// mirror.Process from datastore.
type Store interface {
	// CreateConfig stores a new config in the store.
	CreateConfig(ctx context.Context, config *Config) (*Config, error)

	// UpdateConfig updates a config in the store.
	UpdateConfig(ctx context.Context, config *Config) (*Config, error)

	// GetConfig gets config by namespaceID from the store.
	GetConfig(ctx context.Context, namespaceID uuid.UUID) (*Config, error)

	// CreateProcess stores a new process in the store.
	CreateProcess(ctx context.Context, process *Process) (*Process, error)

	// UpdateProcess update a process in the store.
	UpdateProcess(ctx context.Context, process *Process) (*Process, error)

	// GetProcess gets a process by id from the store.
	GetProcess(ctx context.Context, id uuid.UUID) (*Process, error)

	// GetProcessesByNamespaceID gets all processes that belong to a namespace from the store.
	GetProcessesByNamespaceID(ctx context.Context, namespaceID uuid.UUID) ([]*Process, error)

	// TODO: this need to be refactored.
	SetVariable(ctx context.Context, variable *core.RuntimeVariable) error
}

// Manager launches and terminates mirroring processes.
type Manager interface {
	StartInitialMirroringProcess(ctx context.Context, config *Config) (*Process, error)
	StartSyncingMirrorProcess(ctx context.Context, config *Config) (*Process, error)
	CancelMirroringProcess(ctx context.Context, id uuid.UUID) error
}

// ConfigureWorkflowFunc is a hookup function the gets called for every new or updated workflow file.
type ConfigureWorkflowFunc func(ctx context.Context, file *filestore.File) error

// LogFunc is a hookup function the gets called to perform application logging.
type LogFunc func(processID uuid.UUID, msg string, keysAndValues ...interface{})

// DefaultManager launches and terminates mirroring processes. When launching a mirroring process, DefaultManager
// creates stores and update process objects in the mirror.Store.
type DefaultManager struct {
	infoLogFunc LogFunc
	errLogFunc  LogFunc

	// store is needed so that DefaultManager can create and update mirror.Process objects.
	store Store

	// fStore is to create mirrored files in the filestore.
	fStore filestore.FileStore

	// source is the source of the mirror. Typically, source is a git source.
	source Source

	// configWorkflowFunc is a hookup function the gets called for every new or updated workflow file.
	configWorkflowFunc ConfigureWorkflowFunc
}

func NewDefaultManager(
	infoLogFunc LogFunc,
	errLogFunc LogFunc,
	store Store,
	fStore filestore.FileStore,
	source Source,
	configWorkflowFunc ConfigureWorkflowFunc,
) *DefaultManager {
	if infoLogFunc == nil {
		infoLogFunc = func(processID uuid.UUID, msg string, keysAndValues ...interface{}) {}
	}
	if errLogFunc == nil {
		errLogFunc = func(processID uuid.UUID, msg string, keysAndValues ...interface{}) {}
	}

	return &DefaultManager{
		infoLogFunc:        infoLogFunc,
		errLogFunc:         errLogFunc,
		store:              store,
		fStore:             fStore,
		source:             source,
		configWorkflowFunc: configWorkflowFunc,
	}
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

	d.infoLogFunc(process.ID, "starting mirroring process",
		"type", processType, "process_id", process.ID)

	go func() {
		err := (&mirroringJob{
			ctx:         context.TODO(),
			infoLogFunc: d.infoLogFunc,
		}).
			SetProcessID(process.ID).
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
			process.EndedAt = time.Now()
			process, _ = d.store.UpdateProcess(context.TODO(), process)
			d.errLogFunc(process.ID, "mirroring process failed", "err", err, "process_id", process.ID)

			return
		}

		d.infoLogFunc(process.ID, "mirroring process succeeded", "process_id", process.ID)
	}()

	return process, err
}

// StartInitialMirroringProcess starts an initial mirroring process. The new launched process object is returned.
func (d *DefaultManager) StartInitialMirroringProcess(ctx context.Context, config *Config) (*Process, error) {
	return d.startMirroringProcess(ctx, config, processTypeInit)
}

// StartSyncingMirrorProcess starts a re-mirroring process. The launched process object is returned.
func (d *DefaultManager) StartSyncingMirrorProcess(ctx context.Context, config *Config) (*Process, error) {
	return d.startMirroringProcess(ctx, config, processTypeSync)
}

// nolint:revive
// CancelMirroringProcess stops a currently running mirroring process.
func (d *DefaultManager) CancelMirroringProcess(ctx context.Context, processID uuid.UUID) error {
	// TODO, look if this is needed before release.
	return nil
}

var _ Manager = &DefaultManager{}
