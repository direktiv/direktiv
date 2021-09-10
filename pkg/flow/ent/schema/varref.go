package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// VarRef holds the schema definition for the varref entity.
type VarRef struct {
	ent.Schema
}

// Fields of the VarRef.
func (VarRef) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.String("name").Match(VarNameRegex).Optional().Annotations(entgql.OrderField("NAME")),
	}
}

// Edges of the VarRef.
func (VarRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("vardata", VarData.Type).Ref("varrefs").Unique().Required(),
		edge.From("namespace", Namespace.Type).Ref("vars").Unique(),
		edge.From("workflow", Workflow.Type).Ref("vars").Unique(),
		edge.From("instance", Instance.Type).Ref("vars").Unique(),
	}
}
