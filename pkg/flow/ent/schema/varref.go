package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/direktiv/direktiv/pkg/util"
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
		field.String("name").Match(util.VarNameRegex).Optional(),
		field.String("behaviour").Optional(),
		// TODO: check out if Nillable is required here.
		field.UUID("workflow_id", uuid.UUID{}).Nillable().StorageKey("workflow_id"),
	}
}

// Edges of the VarRef.
func (VarRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("vardata", VarData.Type).Ref("varrefs").Unique().Required(),
		edge.From("namespace", Namespace.Type).Ref("vars").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("instance", Instance.Type).Ref("vars").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
