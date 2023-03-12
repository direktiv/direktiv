package mirror

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Settings holds configuration data are needed to create a mirror (pulling mirror credentials, urls, keys
// and any other details).
type Settings struct {
	ID          uuid.UUID
	NamespaceID uuid.UUID

	Typ           string
	Url           string
	PublicKey     string
	PrivateKey    string
	DecryptionKey string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Activity struct {
	ID         uuid.UUID
	SettingsID uuid.UUID

	Status string

	// TODO: ended at?
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store interface {
	CreateSettings(ctx context.Context, settings *Settings) error
	GetSetting(ctx context.Context, id uuid.UUID) (*Settings, error)
	DeleteSetting(ctx context.Context, id uuid.UUID) error

	CreateActivity(ctx context.Context, activity *Activity) error
	GetActivity(ctx context.Context, id uuid.UUID) (*Activity, error)
	DeleteActivity(ctx context.Context, id uuid.UUID) error
}
