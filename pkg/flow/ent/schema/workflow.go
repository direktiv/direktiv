package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Workflow holds the schema definition for the workflow entity.
type Workflow struct {
	ent.Schema
}

// Fields of the Workflow.
func (Workflow) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Bool("live").Default(true),
	}
}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("inode", Inode.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}).Unique(),
		edge.From("namespace", Namespace.Type).Ref("workflows").Unique().Required(),
		edge.To("revisions", Revision.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("refs", Ref.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("instances", Instance.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("routes", Route.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("logs", LogMsg.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("vars", VarRef.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("wfevents", Events.Type),
	}
}
