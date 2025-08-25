package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("email").NotEmpty(),
		field.String("password_hash").Sensitive(),
		field.String("name").NotEmpty(),
		field.Enum("role").Values("ADMIN", "MANAGER", "CLERK"),
		field.Bool("is_active").Default(true),
		field.Time("last_login").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("users").Unique().Required(),
		edge.To("purchase_orders", PurchaseOrder.Type),
		edge.To("transfers", Transfer.Type),
		edge.To("adjustments", Adjustment.Type),
		edge.To("audit_logs", AuditLog.Type),
		edge.To("stock_movements", StockMovement.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email"),
		index.Fields("is_active"),
		index.Edges("tenant", "email").Unique(), // Email unique per tenant
	}
}
