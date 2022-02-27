package schema

import (
	"time"

	"entgo.io/contrib/entgql"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

const DefaultNamespaceConfig = `
{
	"broadcast": {
	  "workflow.create": false,
	  "workflow.update": false,
	  "workflow.delete": false,
	  "directory.create": false,
	  "directory.delete": false,
	  "workflow.variable.create": false,
	  "workflow.variable.update": false,
	  "workflow.variable.delete": false,
	  "namespace.variable.create": false,
	  "namespace.variable.update": false,
	  "namespace.variable.delete": false,
	  "instance.variable.create": false,
	  "instance.variable.update": false,
	  "instance.variable.delete": false,
	  "instance.started": false,
	  "instance.success": false,
	  "instance.failed": false
	}
  }
`

// Namespace holds the schema definition for the Namespace entity.
type Namespace struct {
	ent.Schema
}

// Fields of the Namespace.
func (Namespace) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.String("config").Default(DefaultNamespaceConfig),
		field.String("name").Match(NameRegex).Annotations(entgql.OrderField("NAME")).Unique().NotEmpty().MaxLen(64).MinLen(1),
	}
}

// Edges of the Namespace.
func (Namespace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("inodes", Inode.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("workflows", Workflow.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("instances", Instance.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("logs", LogMsg.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("vars", VarRef.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("cloudevents", CloudEvents.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("namespacelisteners", Events.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
		edge.To("annotations", Annotation.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
