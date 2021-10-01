package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Route holds the schema definition for the route entity.
type Route struct {
	ent.Schema
}

// Fields of the Route.
func (Route) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Int("weight").Immutable(),
	}
}

// Edges of the Route.
func (Route) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).Ref("routes").Unique().Required().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("ref", Ref.Type).Ref("routes").Unique().Required().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
