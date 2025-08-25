package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Transfer struct {
	ent.Schema
}

func (Transfer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("number").Unique().NotEmpty(),
		field.Enum("status").Values(
			"DRAFT",
			"IN_TRANSIT",
			"RECEIVED",
			"CANCELED",
		).Default("DRAFT"),
		field.Text("notes").Optional(),
		field.Time("shipped_at").Optional().Nillable(),
		field.Time("received_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Transfer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("transfers").Unique().Required(),
		edge.From("from_location", Location.Type).Ref("transfers_from").Unique().Required(),
		edge.From("to_location", Location.Type).Ref("transfers_to").Unique().Required(),
		edge.From("created_by", User.Type).Ref("transfers").Unique(),
		edge.To("approved_by", User.Type).Unique(),
		edge.To("lines", TransferLine.Type),
	}
}

func (Transfer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("number"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Edges("tenant", "number").Unique(), // Transfer number unique per tenant
	}
}
