package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NamespaceSecret holds the schema definition for the NamespaceSecret entity.
type NamespaceSecret struct {
	ent.Schema
}

// Fields of the NamespaceSecret.
func (NamespaceSecret) Fields() []ent.Field {
	return []ent.Field{
		field.String("ns"),
		field.String("name"),
		field.Bytes("secret").MaxLen(65536),
	}
}

// Edges of the NamespaceSecret.
func (NamespaceSecret) Edges() []ent.Edge {
	return nil
}

// Indexes of the secret.
func (NamespaceSecret) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ns").Fields("name").
			Unique(),
	}
}
