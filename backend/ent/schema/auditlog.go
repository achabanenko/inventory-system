package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type AuditLog struct {
	ent.Schema
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("action").NotEmpty(),
		field.String("entity").NotEmpty(),
		field.UUID("entity_id", uuid.UUID{}),
		field.JSON("before", map[string]interface{}{}).Optional(),
		field.JSON("after", map[string]interface{}{}).Optional(),
		field.Time("at").Default(time.Now).Immutable(),
	}
}

func (AuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).Ref("audit_logs").Unique().Required(),
		edge.From("user", User.Type).Ref("audit_logs").Unique(),
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entity", "entity_id"),
		index.Fields("at"),
		index.Fields("action"),
	}
}
