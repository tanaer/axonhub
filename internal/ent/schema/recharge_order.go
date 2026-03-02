package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RechargeOrder holds the schema for recharge orders.
type RechargeOrder struct {
	ent.Schema
}

func (RechargeOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (RechargeOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_id").Unique().Comment("订单号"),
		field.Int("user_id").Comment("用户ID"),
		field.Enum("type").Values("recharge", "package").Default("recharge").Comment("订单类型"),
		field.Int64("amount").Comment("订单金额（单位：分）"),
		field.Enum("status").Values("pending", "paid", "failed", "cancelled").Default("pending").Comment("订单状态"),
		field.String("payment_method").Optional().Comment("支付方式"),
		field.String("trade_id").Optional().Nillable().Comment("第三方交易号"),
		field.Int64("paid_amount").Optional().Nillable().Comment("实际支付金额"),
		field.Time("paid_at").Optional().Nillable().Comment("支付时间"),
		field.Time("expires_at").Optional().Nillable().Comment("过期时间"),
		field.String("description").Optional().Comment("订单描述"),
		field.Int("package_id").Optional().Nillable().Comment("套餐ID（如果购买套餐）"),
	}
}

func (RechargeOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id").Unique(),
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
