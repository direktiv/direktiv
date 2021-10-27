package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// VarData holds the schema definition for the vardata entity.
type VarData struct {
	ent.Schema
}

// Fields of the VarData.
func (VarData) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Int("size"),
		field.String("hash"),
		field.Bytes("data"),
		field.String("mime_type").Default("application/json"),
	}
}

// Edges of the VarData.
func (VarData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("varrefs", VarRef.Type),
	}
}
