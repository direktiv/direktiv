package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Services holds the schema definition for the Services entity.
type Services struct {
	ent.Schema
}

// Fields of the Services.
func (Services) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.String("url").NotEmpty(), // url of the service
		field.String("name").NotEmpty(),
		field.String("data").NotEmpty(),
	}
}

func (Services) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("services").Unique().Required().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}

// Indexes of the secret.
func (Services) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("namespace").Unique(),
	}
}
