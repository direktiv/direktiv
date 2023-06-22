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

type Inode struct {
	ID           uuid.UUID `json:"id,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	Name         string    `json:"name,omitempty"`
	Type         string    `json:"type,omitempty"`
	Attributes   []string  `json:"attributes,omitempty"`
	ExtendedType string    `json:"expandedType,omitempty"`
	ReadOnly     bool      `json:"readOnly,omitempty"`
	Namespace    uuid.UUID `json:"namespace,omitempty"`
	Children     []*Inode  `json:"children,omitempty"`
	Parent       uuid.UUID `json:"parent,omitempty"`
	Workflow     uuid.UUID `json:"workflow,omitempty"`
	Mirror       uuid.UUID `json:"mirror,omitempty"`
}

type Ref struct {
	ID        uuid.UUID `json:"id"`
	Immutable bool      `json:"immutable,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Revision  uuid.UUID `json:"revision,omitempty"`
}

type Revision struct {
	ID        uuid.UUID              `json:"id"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	Hash      string                 `json:"hash,omitempty"`
	Source    []byte                 `json:"source,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Workflow  uuid.UUID              `json:"workflow,omitempty"`
}

type Route struct {
	ID     uuid.UUID `json:"id"`
	Weight int       `json:"weight,omitempty"`
	Ref    *Ref      `json:"ref,omitempty"`
}
