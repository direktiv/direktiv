package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkflowEventsWait holds the schema definition for the WorkflowEventsWait entity.
type WorkflowEventsWait struct {
	ent.Schema
}

// Fields of the WorkflowEventsWait.
func (WorkflowEventsWait) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("events", map[string]interface{}{}),
		// field.Int("count"),
		// field.Int("max"),
	}
}

// Edges of the WorkflowEventsWait.
func (WorkflowEventsWait) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflowevent", WorkflowEvents.Type).
			Ref("wfeventswait").
			Required().
			Unique(),
	}
}
