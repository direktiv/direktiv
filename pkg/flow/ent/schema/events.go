package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Events holds the schema definition for the Events entity.
type Events struct {
	ent.Schema
}

// Fields of the Events.
func (Events) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"id"`).Annotations(entgql.OrderField("ID")),
		field.JSON("events", []map[string]interface{}{}),
		field.JSON("correlations", []string{}),
		field.Bytes("signature").Optional(),
		field.Int("count"),
	}
}

// Edges of the Events.
func (Events) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", Workflow.Type).
			Ref("wfevents").
			Unique().Required(),
		edge.To("wfeventswait", EventsWait.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("instance", Instance.Type).
			Ref("eventlisteners").Unique(),
	}
}
