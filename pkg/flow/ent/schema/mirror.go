package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Mirror holds the schema definition for the mirror entity.
type Mirror struct {
	ent.Schema
}

// Fields of the Mirror.
func (Mirror) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid"),
		field.String("url"),
		field.String("ref"),
		field.String("cron"),
		field.String("public_key").StructTag(`json:"publicKey,omitempty"`),
		field.String("private_key"),
		field.String("passphrase"),
		field.String("commit"),
		field.Time("last_sync").Optional().Nillable(),
		field.Time("updated_at").Optional().Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Mirror.
func (Mirror) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("mirrors").Unique().Required().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("inode", Inode.Type).Ref("mirror").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("activities", MirrorActivity.Type),
	}
}

// MirrorActivity holds the schema definition for the mirror entity.
type MirrorActivity struct {
	ent.Schema
}

// Fields of the MirrorActivity.
func (MirrorActivity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid"),
		field.String("type"),
		field.String("status"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("end_at").Optional(),
		field.String("controller").Optional(),
		field.Time("deadline").Optional(),
	}
}

// Edges of the MirrorActivity.
func (MirrorActivity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("mirror_activities").Required().Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.From("mirror", Mirror.Type).Ref("activities").Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("logs", LogMsg.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
