package datastore

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type JSInstance struct {
	ID           uuid.UUID
	Namespace    string
	WorkflowPath string
	WorkflowData string

	Status int

	Input  sql.NullString
	Memory sql.NullString
	Output sql.NullString
	Error  sql.NullString

	CreatedAt time.Time
	UpdatedAt time.Time
	EndedAt   time.Time
}

func (i JSInstance) StatusString() string {
	if i.Status == 0 {
		return "pending"
	}

	return "N/A"
}

type JSInstancesStore interface {
	Create(ctx context.Context, jsInstance *JSInstance) error
	Patch(ctx context.Context, id uuid.UUID, patch map[string]any) error

	GetByID(ctx context.Context, id uuid.UUID) (*JSInstance, error)
}
