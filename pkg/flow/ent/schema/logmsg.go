package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// LogMsg holds the schema definition for the logmsg entity.
type LogMsg struct {
	ent.Schema
}

// Fields of the LogMsg.
func (LogMsg) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Time("t").Annotations(entgql.OrderField("TIMESTAMP")),
		field.String("msg"),
	}
}

// Edges of the LogMsg.
func (LogMsg) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("logs").Unique(),
		edge.From("workflow", Workflow.Type).Ref("logs").Unique(),
		edge.From("instance", Instance.Type).Ref("logs").Unique(),
	}
}
