package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Item struct {
	ent.Schema
}

func (Item) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("sku").Unique().NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("barcode").Optional().Unique().Nillable(),
		field.String("uom").NotEmpty().Comment("Unit of measure"),
		field.Other("cost", decimal.Decimal{}).SchemaType(map[string]string{
			"postgres": "numeric(10,2)",
		}),
		field.Other("price", decimal.Decimal{}).SchemaType(map[string]string{
			"postgres": "numeric(10,2)",
		}),
		field.JSON("attributes", map[string]interface{}{}).Optional(),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Item) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("items").Unique().Required(),
		edge.To("category", Category.Type).Unique(),
		edge.To("inventory_levels", InventoryLevel.Type),
		edge.To("stock_movements", StockMovement.Type),
		edge.To("purchase_order_lines", PurchaseOrderLine.Type),
		edge.To("transfer_lines", TransferLine.Type),
		edge.To("adjustment_lines", AdjustmentLine.Type),
	}
}

func (Item) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sku"),
		index.Fields("barcode"),
		index.Fields("is_active"),
		index.Fields("deleted_at"),
		index.Edges("tenant", "sku").Unique(),     // SKU unique per tenant
		index.Edges("tenant", "barcode").Unique(), // Barcode unique per tenant
	}
}
