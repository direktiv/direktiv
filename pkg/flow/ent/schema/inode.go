package schema

import (
	"time"

	"entgo.io/contrib/entgql"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Inode holds the schema definition for the inode entity.
type Inode struct {
	ent.Schema
}

// Fields of the Inode.
func (Inode) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid"),
		field.Time("created_at").Default(time.Now).Immutable().Annotations(entgql.OrderField("CREATED")),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now).Annotations(entgql.OrderField("UPDATED")),
		field.String("name").Match(NameRegex).Optional().Annotations(entgql.OrderField("NAME")),
		field.String("type").Immutable(),
		field.Strings("attributes").Optional(),
	}
}

// Edges of the Inode.
func (Inode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("inodes").Unique().Required(),
		edge.To("children", Inode.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("parent", Inode.Type).Ref("children").Unique(),
		// edge.From("workflow", Workflow.Type).Ref("inode").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("workflow", Workflow.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}).Unique(),
	}
}

// Indexes of the Inode.
func (Inode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("parent").Unique(),
	}
}
