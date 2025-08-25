package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type StockMovement struct {
	ent.Schema
}

func (StockMovement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int("qty").Comment("Positive for receipts, negative for issues"),
		field.Enum("reason").Values(
			"PO_RECEIPT",
			"ADJUSTMENT",
			"TRANSFER_OUT",
			"TRANSFER_IN",
			"COUNT",
		),
		field.String("reference").Optional(),
		field.UUID("ref_id", uuid.UUID{}).Optional().Nillable(),
		field.JSON("meta", map[string]interface{}{}).Optional(),
		field.Time("occurred_at").Default(time.Now),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (StockMovement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("stock_movements").Unique().Required(),
		edge.From("item", Item.Type).Ref("stock_movements").Unique().Required(),
		edge.From("location", Location.Type).Ref("stock_movements").Unique().Required(),
		edge.From("user", User.Type).Ref("stock_movements").Unique(),
	}
}

func (StockMovement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("occurred_at"),
		index.Fields("reason"),
		index.Fields("ref_id"),
	}
}
