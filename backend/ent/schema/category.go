package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").NotEmpty(),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("categories").Unique().Required(),
		edge.From("items", Item.Type).Ref("category"),
		edge.To("children", Category.Type).From("parent").Unique(),
	}
}

func (Category) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("tenant", "name").Unique(), // Category name unique per tenant
	}
}
