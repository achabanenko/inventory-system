package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type PurchaseOrder struct {
	ent.Schema
}

func (PurchaseOrder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("number").Unique().NotEmpty(),
		field.Enum("status").Values(
			"DRAFT",
			"APPROVED",
			"PARTIAL",
			"RECEIVED",
			"CLOSED",
			"CANCELED",
		).Default("DRAFT"),
		field.Time("expected_at").Optional().Nillable(),
		field.Text("notes").Optional(),
		field.Time("approved_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (PurchaseOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("purchase_orders").Unique().Required(),
		edge.From("supplier", Supplier.Type).Ref("purchase_orders").Unique().Required(),
		edge.From("created_by", User.Type).Ref("purchase_orders").Unique(),
		edge.To("approved_by", User.Type).Unique(),
		edge.To("lines", PurchaseOrderLine.Type),
	}
}

func (PurchaseOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("number"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Edges("tenant", "number").Unique(), // PO number unique per tenant
	}
}
