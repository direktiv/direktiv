package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// CloudEvents holds the schema definition for the events.
type CloudEvents struct {
	ent.Schema
}

// Fields of the CloudEvent.
func (CloudEvents) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Immutable().Unique().NotEmpty(),
		field.String("namespace").Immutable().NotEmpty(),
		field.JSON("event", cloudevents.Event{}),
		field.Time("fire").Immutable().Default(time.Now),
		field.Time("created").Immutable().Default(time.Now),
		field.Bool("processed"),
	}
}
