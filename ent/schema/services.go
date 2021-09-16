package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Services holds the schema definition for the Services entity.
type Services struct {
	ent.Schema
}

// Fields of the Services.
func (Services) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Immutable().Unique().NotEmpty().MinLen(1),
		field.String("data"),
	}
}
