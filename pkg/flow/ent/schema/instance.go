package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Instance holds the schema definition for the instance entity.
type Instance struct {
	ent.Schema
}

// Fields of the Instance.
func (Instance) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"id"`),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("end_at").Optional(),
		field.String("status"),
		field.String("as").Immutable(),
		field.String("errorCode").Optional(),
		field.String("errorMessage").Optional(),
		field.String("invoker").Optional(),
		field.String("invokerState").Optional(),
		field.String("callpath").Optional(),
		// TODO: check out if Nillable is required here.
		field.UUID("workflow_id", uuid.UUID{}).Nillable().StorageKey("workflow_id"),
		// TODO: check out if Nillable is required here.
		field.UUID("revision_id", uuid.UUID{}).Nillable().StorageKey("revision_id"),
	}
}

// Edges of the Instance.
func (Instance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("namespace", Namespace.Type).Ref("instances").Required().Unique().Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("logs", LogMsg.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("vars", VarRef.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("runtime", InstanceRuntime.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}).Unique().Required(),
		edge.To("children", InstanceRuntime.Type),
		edge.To("eventlisteners", Events.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("annotations", Annotation.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
