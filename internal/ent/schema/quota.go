package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/looplj/axonhub/internal/ent/schema/schematype"
)

// UserQuota holds the schema definition for the UserQuota entity.
type UserQuota struct {
	ent.Schema
}

func (UserQuota) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (UserQuota) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Unique().Comment("关联的用户ID"),
		field.Int64("quota").Default(0).Comment("账户余额（单位：分）"),
		field.Int64("used_quota").Default(0).Comment("已使用额度（单位：分）"),
		field.String("group").Default("default").Comment("用户分组/等级"),
		field.String("aff_code").Optional().Nillable().Comment("邀请码"),
		field.Int("inviter_id").Optional().Nillable().Comment("邀请人ID"),
		field.Bool("notify").Default(true).Comment("是否接收余额提醒"),
		field.Int64("quota_remind_threshold").Default(100000).Comment("余额提醒阈值（单位：分，默认20元）"),
	}
}

func (UserQuota) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("aff_code").Unique(),
		index.Fields("group"),
	}
}

// QuotaTransaction holds the schema for quota transaction records.
type QuotaTransaction struct {
	ent.Schema
}

func (QuotaTransaction) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (QuotaTransaction) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Comment("用户ID"),
		field.Enum("type").Values("recharge", "consume", "refund", "reward").Comment("交易类型"),
		field.Int64("amount").Comment("金额（单位：分）"),
		field.Int64("balance_after").Comment("交易后余额"),
		field.String("description").Optional().Comment("交易描述"),
		field.String("order_id").Optional().Comment("关联订单号"),
		field.Enum("status").Values("pending", "completed", "failed").Default("completed"),
	}
}

// SubscriptionPlan holds the schema for subscription plans.
type SubscriptionPlan struct {
	ent.Schema
}

func (SubscriptionPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		schematype.SoftDeleteMixin{},
	}
}

func (SubscriptionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().Comment("套餐名称"),
		field.String("display_name").Comment("显示名称"),
		field.Int64("price").Comment("价格（单位：分）"),
		field.String("currency").Default("CNY").Comment("货币"),
		field.Int64("quota").Comment("包含额度（单位：分）"),
		field.Int64("duration_days").Comment("有效天数"),
		field.Strings("features").Optional().Comment("套餐特性列表"),
		field.Bool("is_popular").Default(false).Comment("是否为推荐套餐"),
		field.Int("sort_order").Default(0).Comment("排序"),
		field.Bool("status").Default(true).Comment("是否启用"),
	}
}

// UserSubscription holds user's active subscriptions.
type UserSubscription struct {
	ent.Schema
}

func (UserSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

func (UserSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Comment("用户ID"),
		field.Int("plan_id").Comment("套餐ID"),
		field.Int64("quota_remaining").Comment("剩余额度"),
		field.Time("expires_at").Comment("过期时间"),
		field.Enum("status").Values("active", "expired", "cancelled").Default("active"),
		field.String("order_id").Optional().Comment("订单号"),
	}
}
