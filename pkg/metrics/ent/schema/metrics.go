package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Metrics holds the schema definition for the Metrics entity.
type Metrics struct {
	ent.Schema
}

// Fields of the Metrics.
func (Metrics) Fields() []ent.Field {
	return []ent.Field{
		// field.String("id").Unique(),
		field.String("namespace").NotEmpty(),
		field.String("workflow").NotEmpty(),
		field.String("revision"),
		field.String("instance").NotEmpty(),
		field.String("state").NotEmpty(),
		field.Time("timestamp"),
		field.Int64("workflow_ms").NonNegative(),
		field.Int64("isolate_ms").NonNegative(),
		field.String("error_code").Optional(),
		field.String("invoker"),
		field.Int8("next").Min(0).Max(2),
		field.String("transition").Optional(),
	}
}

// Edges of the Metrics.
func (Metrics) Edges() []ent.Edge {
	return nil
}
