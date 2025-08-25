package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Tenant struct {
	ent.Schema
}

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty().Comment("Tenant/Company name"),
		field.String("slug").Unique().NotEmpty().Comment("URL-safe identifier"),
		field.String("domain").Optional().Unique().Nillable().Comment("Custom domain for tenant"),
		field.JSON("settings", map[string]interface{}{}).Optional().Comment("Tenant-specific settings"),
		field.JSON("contact", map[string]interface{}{}).Optional().Comment("Tenant contact information"),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
		edge.To("items", Item.Type),
		edge.To("categories", Category.Type),
		edge.To("locations", Location.Type),
		edge.To("suppliers", Supplier.Type),
		edge.To("inventory_levels", InventoryLevel.Type),
		edge.To("stock_movements", StockMovement.Type),
		edge.To("purchase_orders", PurchaseOrder.Type),
		edge.To("transfers", Transfer.Type),
		edge.To("adjustments", Adjustment.Type),
		edge.To("audit_logs", AuditLog.Type),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug"),
		index.Fields("domain"),
		index.Fields("is_active"),
	}
}
