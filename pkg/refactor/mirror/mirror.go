package mirror

import (
	"context"
	"time"

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
