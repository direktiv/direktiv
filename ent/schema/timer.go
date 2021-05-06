package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Timer holds the schema definition for the Cron entity.
type Timer struct {
	ent.Schema
}

// Fields of the Cron.
func (Timer) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().Immutable().NotEmpty(),
		field.String("fn").Immutable().NotEmpty(),
		field.String("cron").Immutable().Optional(),
		field.Time("one").Immutable().Optional(),
		field.Bytes("data").Immutable().Optional(),
		field.Time("last").Optional(),
	}
}
