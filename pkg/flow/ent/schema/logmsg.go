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
		field.String("level").Default("info"),
		field.String("rootInstanceId").Default(""), // NOTE: this field is redundant, but it allows us to improve query performance.
		field.String("logInstanceCallPath").Default(""),
		field.JSON("tags", map[string]string{}).Optional(),
		// TODO: check out if Nillable is required here.
		field.UUID("workflow_id", uuid.UUID{}).Optional().StorageKey("workflow_id"),
		// TODO: check out if Nillable is required here.
		field.UUID("mirror_activity_id", uuid.UUID{}).Optional().StorageKey("mirror_activity_id"),
		field.UUID("instance_id", uuid.UUID{}).Optional().StorageKey("instance_id"),
	}
}

// Edges of the LogMsg.
func (LogMsg) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("logs").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		// edge.From("instance", Instance.Type).Ref("logs").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
