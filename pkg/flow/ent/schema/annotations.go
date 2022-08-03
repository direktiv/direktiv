package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/direktiv/direktiv/pkg/util"
	"github.com/google/uuid"
)

// Annotation holds the schema definition for the annotation entity.
type Annotation struct {
	ent.Schema
}

// Fields of the Annotation.
func (Annotation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.String("name").Match(util.VarNameRegex),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Int("size"),
		field.String("hash"),
		field.Bytes("data"),
	}
}

// Edges of the Annotation.
func (Annotation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("annotations").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("workflow", Workflow.Type).Ref("annotations").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("instance", Instance.Type).Ref("annotations").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("inode", Inode.Type).Ref("annotations").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
