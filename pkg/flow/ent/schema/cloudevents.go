package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// CloudEvents holds the schema definition for the events.
type CloudEvents struct {
	ent.Schema
}

// Fields of the CloudEvent.
func (CloudEvents) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"id"`).Annotations(entgql.OrderField("ID")),
		field.String("eventId").Immutable().Unique().NotEmpty(),
		field.JSON("event", cloudevents.Event{}),
		field.Time("fire").Immutable().Default(time.Now),
		field.Time("created").Immutable().Default(time.Now),
		field.Bool("processed"),
	}
}

// Edges of the CloudEvents.
func (CloudEvents) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("cloudevents").Unique().Required(),
	}
}
