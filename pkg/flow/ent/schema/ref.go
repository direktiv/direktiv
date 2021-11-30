package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Ref holds the schema definition for the ref entity.
type Ref struct {
	ent.Schema
}

// Fields of the Ref.
func (Ref) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Bool("immutable").Default(true).Immutable(),
		field.String("name").Match(RefRegex).Immutable().Annotations(entgql.OrderField("NAME")),
		field.Time("created_at").Default(time.Now).Immutable().Annotations(entgql.OrderField("CREATED")),
	}
}

// Edges of the Ref.
func (Ref) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).Ref("refs").Unique().Required().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("revision", Revision.Type).Ref("refs").Unique().Required().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("routes", Route.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}

// Indexes of the Ref.
func (Ref) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("workflow").Unique(),
	}
}
