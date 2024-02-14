package mirror

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Config holds configuration data that are needed to create a mirror (pulling mirror credentials, urls, keys
// and any other details).
type Config struct {
	Namespace string

	URL                  string
	GitRef               string
	GitCommitHash        string
	PublicKey            string
	PrivateKey           string
	PrivateKeyPassphrase string

	CreatedAt time.Time
	UpdatedAt time.Time

	Insecure bool
}

// Process different statuses.
const (
	ProcessStatusComplete  = "complete"
	ProcessStatusPending   = "pending"
	ProcessStatusExecuting = "executing"
	ProcessStatusFailed    = "failed"
)

// Process different types.
const (
	// Indicates initial mirroring process.
	ProcessTypeInit = "init"

	// Indicates re-mirroring process.
	ProcessTypeSync = "sync"

	// Indicates dry run process.
	ProcessTypeDryRun = "dryrun"
)

// Process represents an instance of mirroring process that happened or is currently happened. For every mirroring
// process gets executing, a Process instance should be created with mirror.Store.
type Process struct {
	ID        uuid.UUID
	Namespace string

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

	// GetConfig gets config by namespace from the store.
	GetConfig(ctx context.Context, namespace string) (*Config, error)

	GetAllConfigs(ctx context.Context) ([]*Config, error)
	// CreateProcess stores a new process in the store.
	CreateProcess(ctx context.Context, process *Process) (*Process, error)

	// UpdateProcess update a process in the store.
	UpdateProcess(ctx context.Context, process *Process) (*Process, error)

	// GetProcess gets a process by id from the store.
	GetProcess(ctx context.Context, id uuid.UUID) (*Process, error)

	// GetProcessesByNamespace gets all processes that belong to a namespace from the store.
	GetProcessesByNamespace(ctx context.Context, namespace string) ([]*Process, error)

	// GetUnfinishedProcesses gets all processes that haven't completed from the store.
	GetUnfinishedProcesses(ctx context.Context) ([]*Process, error)

	// DeleteOldProcesses deletes all old processes.
	DeleteOldProcesses(ctx context.Context, before time.Time) error
}
