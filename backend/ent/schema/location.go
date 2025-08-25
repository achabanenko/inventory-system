package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Location struct {
	ent.Schema
}

func (Location) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("code").Unique().NotEmpty(),
		field.String("name").NotEmpty(),
		field.JSON("address", map[string]interface{}{}).Optional(),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Location) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("locations").Unique().Required(),
		edge.To("inventory_levels", InventoryLevel.Type),
		edge.To("stock_movements", StockMovement.Type),
		edge.To("transfers_from", Transfer.Type),
		edge.To("transfers_to", Transfer.Type),
		edge.To("adjustments", Adjustment.Type),
	}
}

func (Location) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code"),
		index.Fields("is_active"),
		index.Edges("tenant", "code").Unique(), // Location code unique per tenant
	}
}
