package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CloudEvents holds the schema definition for the events.
type CloudEventFilters struct {
	ent.Schema
}

// Fields of the CloudEvent.
func (CloudEventFilters) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("jscode").NotEmpty(),
	}
}

// Edges of the CloudEventFilters.
func (CloudEventFilters) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("cloudeventfilters").Unique().Required(),
	}
}

// Indexes of the CloudEventFilters.
func (CloudEventFilters) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("namespace").Unique(),
	}
}
