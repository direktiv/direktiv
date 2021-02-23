package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// BucketSecret holds the schema definition for the BucketSecret entity.
type BucketSecret struct {
	ent.Schema
}

// Fields of the BucketSecret.
func (BucketSecret) Fields() []ent.Field {
	return []ent.Field{
		field.String("ns"),
		field.String("name"),
		field.Bytes("secret").MaxLen(65536),
		field.Int("type").Default(0),
	}
}

// // Edges of the BucketSecret.
// func (BucketSecret) Edges() []ent.Edge {
// 	return []ent.Edge{
// 		edge.From("bucket", Bucket.Type).
// 			Ref("secret").
// 			Unique().Required(),
// 	}
// }
