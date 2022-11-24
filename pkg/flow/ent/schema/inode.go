package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/direktiv/direktiv/pkg/util"
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
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.String("name").Match(util.NameRegex).Optional(),
		field.String("type").Immutable(),
		field.Strings("attributes").Optional(),
		field.String("extended_type").Optional().StorageKey("expandedType").StructTag(`json:"expandedType,omitempty"`),
		field.Bool("readOnly").Optional().Default(false),
	}
}

// Edges of the Inode.
func (Inode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("inodes").Unique().Required(),
		edge.To("children", Inode.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("parent", Inode.Type).Ref("children").Unique(),
		edge.To("workflow", Workflow.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}).Unique(),
		edge.To("mirror", Mirror.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}).Unique(),
		edge.To("annotations", Annotation.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}

// Indexes of the Inode.
func (Inode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("parent").Unique(),
	}
}
