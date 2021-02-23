package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkflowEvents holds the schema definition for the WorkflowEvents entity.
type WorkflowEvents struct {
	ent.Schema
}

// Fields of the WorkflowEvents.
func (WorkflowEvents) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("events", []map[string]interface{}{}),
		field.JSON("correlations", []string{}),
		field.Bytes("signature").Optional(),
		field.Int("count"),
	}
}

// Edges of the WorkflowEvents.
func (WorkflowEvents) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).
			Ref("wfevents").
			Unique().Required(),
		edge.To("wfeventswait", WorkflowEventsWait.Type),
	}
}
