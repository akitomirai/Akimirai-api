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

// ExternalFulfillmentSKU maps marketplace SKUs to redeem-code packages.
type ExternalFulfillmentSKU struct {
	ent.Schema
}

func (ExternalFulfillmentSKU) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "external_fulfillment_skus"},
	}
}

func (ExternalFulfillmentSKU) Fields() []ent.Field {
	return []ent.Field{
		field.String("platform").
			MaxLen(32).
			Default("xianyu"),
		field.String("sku_code").
			MaxLen(128).
			NotEmpty(),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}).
			Default(0),
		field.String("currency").
			MaxLen(16).
			Default("CNY"),
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
		field.Int("expires_in_days").
			Optional().
			Nillable(),
		field.String("manual_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("delivery_template").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Bool("enabled").
			Default(true),
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

func (ExternalFulfillmentSKU) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "sku_code").Unique(),
		index.Fields("platform", "enabled"),
	}
}
