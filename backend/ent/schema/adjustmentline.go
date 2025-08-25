package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type AdjustmentLine struct {
	ent.Schema
}

func (AdjustmentLine) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int("qty_diff").Comment("Positive or negative adjustment"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (AdjustmentLine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("adjustment", Adjustment.Type).Ref("lines").Unique().Required(),
		edge.From("item", Item.Type).Ref("adjustment_lines").Unique().Required(),
	}
}