package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Services holds the schema definition for the Services entity.
type Services struct {
	ent.Schema
}

// Fields of the Services.
func (Services) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"), // url of the service
		field.String("name").Unique().NotEmpty(),
		field.String("data"),
	}
}

func (Services) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("services").Unique().Required().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
