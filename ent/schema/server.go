package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Server holds the schema definition for the Server entity.
type Server struct {
	ent.Schema
}

// Fields of the Server.
func (Server) Fields() []ent.Field {
	return []ent.Field{
		field.String("ip").Immutable().Unique().NotEmpty(),
		field.String("extIP").Immutable().NotEmpty(),
		field.Int("natsPort").Immutable(),
		field.Int("memberPort").Immutable(),
		field.Time("added").Immutable().Default(time.Now),
	}
}

// Edges of the Server.
func (Server) Edges() []ent.Edge {
	return nil
}
