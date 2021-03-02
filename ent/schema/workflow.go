package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Workflow holds the schema definition for th e Workflow entity.
type Workflow struct {
	ent.Schema
}

// Fields of the Workflow.
func (Workflow) Fields() []ent.Field {

	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty(),
		field.Time("created").Immutable().Default(time.Now),
		field.String("description").Optional().MaxLen(1024).Default(""),
		field.Bool("active").Default(true),
		field.Int("revision").Default(0),
		field.Bytes("workflow"),
	}

}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).
			Ref("workflows").
			Unique().Required(),
		edge.To("instances", WorkflowInstance.Type),
		edge.To("wfevents", WorkflowEvents.Type),
	}
}

// Indexes of the Workflow.
func (Workflow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("namespace").
			Unique(),
	}
}
