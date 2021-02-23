package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Subroutine holds the schema definition for the Subroutine entity.
type Subroutine struct {
	ent.Schema
}

// Fields of the Subroutine.
func (Subroutine) Fields() []ent.Field {
	return []ent.Field{
		field.String("callerID").Unique(),
		field.Int("semaphore"),
		field.String("memory"),
		field.Strings("subroutineIDs"),
		field.Strings("subroutineResponses").Optional(),
	}
}

// Edges of the Subroutine.
func (Subroutine) Edges() []ent.Edge {
	return nil
}
