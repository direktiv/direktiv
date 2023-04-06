package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// InstanceRuntime holds the schema definition for the instance runtime entity.
type InstanceRuntime struct {
	ent.Schema
}

// Fields of the InstanceRuntime.
func (InstanceRuntime) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Bytes("input").Immutable(),
		field.String("data"),
		field.String("controller").Optional(),
		field.String("memory").Optional(),
		field.Strings("flow").Optional(),
		field.String("output").Optional(),
		field.Time("stateBeginTime").Optional(),
		field.Time("deadline").Optional(),
		field.Int("attempts").Optional(),
		field.String("caller_data").Optional(),
		field.String("instanceContext").Optional(),
		field.String("stateContext").Optional(),
		field.String("metadata").Optional(),
		field.String("logToEvents").Optional(),
	}
}

// Edges of the InstanceRuntime.
func (InstanceRuntime) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", Instance.Type).Ref("runtime").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("caller", Instance.Type).Ref("children").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
