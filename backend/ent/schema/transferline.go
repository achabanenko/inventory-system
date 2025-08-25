package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type TransferLine struct {
	ent.Schema
}

func (TransferLine) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int("qty").Min(1),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (TransferLine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("transfer", Transfer.Type).Ref("lines").Unique().Required(),
		edge.From("item", Item.Type).Ref("transfer_lines").Unique().Required(),
	}
}