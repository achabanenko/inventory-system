package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PurchaseOrderLine struct {
	ent.Schema
}

func (PurchaseOrderLine) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int("qty_ordered").Min(1),
		field.Int("qty_received").Default(0).Min(0),
		field.Other("unit_cost", decimal.Decimal{}).SchemaType(map[string]string{
			"postgres": "numeric(10,2)",
		}),
		field.JSON("tax", map[string]interface{}{}).Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PurchaseOrderLine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("purchase_order", PurchaseOrder.Type).Ref("lines").Unique().Required(),
		edge.From("item", Item.Type).Ref("purchase_order_lines").Unique().Required(),
	}
}