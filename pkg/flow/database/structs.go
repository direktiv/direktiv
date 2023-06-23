package database

import (
	"time"

	"github.com/google/uuid"
)

type Namespace struct {
	ID        uuid.UUID `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Config    string    `json:"config,omitempty"`
	Name      string    `json:"name,omitempty"`
	Root      uuid.UUID `json:"root,omitempty"`
}

func (ns *Namespace) GetAttributes() map[string]string {
	return map[string]string{
		"namespace":    ns.Name,
		"namespace-id": ns.ID.String(),
	}
}
