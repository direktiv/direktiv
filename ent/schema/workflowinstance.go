package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// WorkflowInstance holds the schema definition for the WorkflowInstance entity.
type WorkflowInstance struct {
	ent.Schema
}

// Fields of the WorkflowInstance.
func (WorkflowInstance) Fields() []ent.Field {
	return []ent.Field{
		field.String("instanceID").Unique(),
		field.String("invokedBy"),
		field.String("status"),
		field.Int("revision"),
		field.Time("beginTime"),
		field.Time("endTime").Optional(),
		field.Strings("flow").Optional(),
		field.String("input"),
		field.String("output").Optional(),
		field.String("stateData").Optional(),
		field.String("memory").Optional(),
		field.Time("deadline").Optional(),
		field.Int("attempts").Optional(),
		field.String("errorCode").Optional(),
		field.String("errorMessage").Optional(),
		field.Time("stateBeginTime").Optional(),
	}
}

// Edges of the WorkflowInstance.
func (WorkflowInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).
			Ref("instances").
			Unique().Required(),
		edge.To("instance", WorkflowEvents.Type),
	}
}
