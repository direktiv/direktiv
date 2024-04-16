package datastore

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

// MirrorConfig holds configuration data that are needed to create a mirror (pulling mirror credentials, urls, keys
// and any other details).
type MirrorConfig struct {
	Namespace string

	URL                  string
	GitRef               string
	AuthType             string
	AuthToken            string
	PublicKey            string
	PrivateKey           string
	PrivateKeyPassphrase string
	Insecure             bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("validation errors: %v", v)
}

var _ error = ValidationError{}

func (m *MirrorConfig) Validate() ValidationError {
	result := map[string]string{}

	if m.Namespace == "" {
		result["namespace"] = "field is required"
	}
	if m.URL == "" {
		result["url"] = "field is required"
	}
	if m.GitRef == "" {
		result["gitRef"] = "field is required"
	}
	if !slices.Contains([]string{"public", "ssh", "token", ""}, m.AuthType) {
		result["authType"] = "has not allowed enum value"
	}
	if m.AuthType == "token" && m.AuthToken == "" {
		result["authToken"] = "should not be empty with authType=token"
	}
	if m.AuthType == "ssh" && m.PublicKey == "" {
		result["publicKey"] = "should not be empty with authType=ssh"
	}
	if m.AuthType == "ssh" && m.PrivateKey == "" {
		result["privateKey"] = "should not be empty with authType=ssh"
	}

	return result
}

// MirrorProcess different statuses.
const (
	ProcessStatusComplete  = "complete"
	ProcessStatusPending   = "pending"
	ProcessStatusExecuting = "executing"
	ProcessStatusFailed    = "failed"
)

// MirrorProcess different types.
const (
	// Indicates initial mirroring process.
	ProcessTypeInit = "init"

	// Indicates re-mirroring process.
	ProcessTypeSync = "sync"

	// Indicates dry run process.
	ProcessTypeDryRun = "dryrun"
)

// MirrorProcess represents an instance of mirroring process that happened or is currently happened. For every mirroring
// process gets executing, a MirrorProcess instance should be created with datastore.MirrorStore.
type MirrorProcess struct {
	ID        uuid.UUID `json:"id"`
	Namespace string    `json:"-"`
	Status    string    `json:"status"`
	Typ       string    `json:"-"`
	EndedAt   time.Time `json:"endedAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MirrorStore *doesn't* lunch any mirroring process. MirrorStore is only responsible for fetching and setting datastore.MirrorConfig and
// datastore.MirrorProcess from datastore.
// nolint: interfacebloat
type MirrorStore interface {
	// CreateConfig stores a new config in the store.
	CreateConfig(ctx context.Context, config *MirrorConfig) (*MirrorConfig, error)

	// UpdateConfig updates a config in the store.
	UpdateConfig(ctx context.Context, config *MirrorConfig) (*MirrorConfig, error)

	// GetConfig gets config by namespace from the store.
	GetConfig(ctx context.Context, namespace string) (*MirrorConfig, error)

	GetAllConfigs(ctx context.Context) ([]*MirrorConfig, error)

	// DeleteConfig deletes mirror config of a namespace
	DeleteConfig(ctx context.Context, namespace string) error

	// CreateProcess stores a new process in the store.
	CreateProcess(ctx context.Context, process *MirrorProcess) (*MirrorProcess, error)

	// UpdateProcess update a process in the store.
	UpdateProcess(ctx context.Context, process *MirrorProcess) (*MirrorProcess, error)

	// GetProcess gets a process by id from the store.
	GetProcess(ctx context.Context, id uuid.UUID) (*MirrorProcess, error)

	// GetProcessesByNamespace gets all processes that belong to a namespace from the store.
	GetProcessesByNamespace(ctx context.Context, namespace string) ([]*MirrorProcess, error)

	// GetUnfinishedProcesses gets all processes that haven't completed from the store.
	GetUnfinishedProcesses(ctx context.Context) ([]*MirrorProcess, error)

	// DeleteOldProcesses deletes all old processes.
	DeleteOldProcesses(ctx context.Context, before time.Time) error
}
