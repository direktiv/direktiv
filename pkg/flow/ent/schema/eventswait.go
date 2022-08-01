package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// EventsWait holds the schema definition for the EventsWait entity.
type EventsWait struct {
	ent.Schema
}

// Fields of the EventsWait.
func (EventsWait) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"id"`),
		field.JSON("events", map[string]interface{}{}),
	}
}

// Edges of the EventsWait.
func (EventsWait) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflowevent", Events.Type).Ref("wfeventswait").Required().Unique(),
	}
}
