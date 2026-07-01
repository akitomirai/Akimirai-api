package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ExternalOrderFulfillment stores marketplace order delivery state.
type ExternalOrderFulfillment struct {
	ent.Schema
}

func (ExternalOrderFulfillment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "external_order_fulfillments"},
	}
}

func (ExternalOrderFulfillment) Fields() []ent.Field {
	return []ent.Field{
		field.String("platform").
			MaxLen(32).
			Default("xianyu"),
		field.String("platform_order_id").
			MaxLen(128).
			NotEmpty(),
		field.String("buyer_ref").
			Optional().
			Nillable().
			MaxLen(255),
		field.String("sku_code").
			MaxLen(128).
			NotEmpty(),
		field.String("sku_name").
			Optional().
			Nillable().
			MaxLen(255),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.String("currency").
			MaxLen(16).
			Default("CNY"),
		field.Int64("redeem_code_id").
			Optional().
			Nillable(),
		field.String("redeem_code").
			Optional().
			Nillable().
			MaxLen(128),
		field.String("redeem_type").
			MaxLen(32),
		field.Float("redeem_value").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,6)"}).
			Default(0),
		field.Int64("group_id").
			Optional().
			Nillable(),
		field.Int("validity_days").
			Default(0),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("manual_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("delivery_message").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("status").
			MaxLen(32).
			Default("pending"),
		field.String("notify_status").
			MaxLen(32).
			Default("skipped"),
		field.String("fail_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("operator").
			Optional().
			Nillable().
			MaxLen(128),
		field.Time("delivered_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("notified_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ExternalOrderFulfillment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "platform_order_id").Unique(),
		index.Fields("status", "created_at"),
		index.Fields("platform", "sku_code", "created_at"),
		index.Fields("redeem_code_id"),
	}
}
