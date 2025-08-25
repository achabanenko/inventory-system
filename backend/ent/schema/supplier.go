package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Supplier struct {
	ent.Schema
}

func (Supplier) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("code").Unique().NotEmpty(),
		field.String("name").NotEmpty(),
		field.JSON("contact", map[string]interface{}{}).Optional(),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Supplier) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("suppliers").Unique().Required(),
		edge.To("purchase_orders", PurchaseOrder.Type),
	}
}

func (Supplier) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code"),
		index.Fields("is_active"),
		index.Edges("tenant", "code").Unique(), // Supplier code unique per tenant
	}
}
