package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Adjustment struct {
	ent.Schema
}

func (Adjustment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("number").Unique().NotEmpty(),
		field.Enum("reason").Values(
			"COUNT",
			"DAMAGE",
			"CORRECTION",
		),
		field.Enum("status").Values(
			"DRAFT",
			"APPROVED",
			"CANCELED",
		).Default("DRAFT"),
		field.Text("notes").Optional(),
		field.Time("approved_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Adjustment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("adjustments").Unique().Required(),
		edge.From("location", Location.Type).Ref("adjustments").Unique().Required(),
		edge.From("created_by", User.Type).Ref("adjustments").Unique(),
		edge.To("approved_by", User.Type).Unique(),
		edge.To("lines", AdjustmentLine.Type),
	}
}

func (Adjustment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("number"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Edges("tenant", "number").Unique(), // Adjustment number unique per tenant
	}
}
