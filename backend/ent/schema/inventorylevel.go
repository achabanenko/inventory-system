package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type InventoryLevel struct {
	ent.Schema
}

func (InventoryLevel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int("on_hand").Default(0).Min(0),
		field.Int("allocated").Default(0).Min(0),
		field.Int("reorder_point").Default(0).Min(0),
		field.Int("reorder_qty").Default(0).Min(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (InventoryLevel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("inventory_levels").Unique().Required(),
		edge.From("item", Item.Type).Ref("inventory_levels").Unique().Required(),
		edge.From("location", Location.Type).Ref("inventory_levels").Unique().Required(),
	}
}

func (InventoryLevel) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tenant", "item", "location").Unique(), // One inventory level per item per location per tenant
	}
}
