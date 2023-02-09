package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// LogMsg holds the schema definition for the logmsg entity.
type LogMsg struct {
	ent.Schema
}

// Fields of the LogMsg.
func (LogMsg) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.Time("t"),
		field.String("msg"),
	}
}

// Edges of the LogMsg.
func (LogMsg) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("logs").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("workflow", Workflow.Type).Ref("logs").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("instance", Instance.Type).Ref("logs").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("activity", MirrorActivity.Type).Ref("logs").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("logtag", LogTag.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
