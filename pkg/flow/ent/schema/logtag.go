package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// LogMsg holds the schema definition for the logmsg entity.
type LogTag struct {
	ent.Schema
}

// Fields of the LogMsg.
func (LogTag) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable().StorageKey("oid").StructTag(`json:"-"`),
		field.String("type"),
		field.String("value"),
	}
}

func (LogTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("logmsg", LogMsg.Type).Ref("logtag").Immutable().Unique().Required(),
	}
}
