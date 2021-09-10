package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Revision holds the schema definition for the revision entity.
type Revision struct {
	ent.Schema
}

// Fields of the Revision.
func (Revision) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.String("hash").Immutable(),
		field.Bytes("source").Immutable(),
	}
}

// Edges of the Revision.
func (Revision) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).Ref("revisions").Unique().Required(),
		edge.To("refs", Ref.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("instances", Instance.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
