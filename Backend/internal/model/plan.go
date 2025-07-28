package model

import (
	"academy/internal/types"
	"time"

	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type PlanID string

const (
	PlanIDFreeForever     = "free_forever"
	PlanIDStandardMonthly = "standard_monthly"
	PlanIDStandardYearly  = "standard_yearly"
	PlanIDPremiumMonthly  = "premium_monthly"
	PlanIDPremiumYearly   = "premium_yearly"
	PlanIDPremiumPromo    = "premium_promo"
)

const DefaultPlanID PlanID = PlanIDPremiumPromo

type Plan struct {
	bun.BaseModel `bun:"table:plans"`

	ID          PlanID          `bun:"id,pk,type:varchar(100)" json:"id"`
	Name        string          `bun:"name,type:varchar(100),notnull" json:"name"`
	Description string          `bun:"description,type:text,notnull" json:"description"`
	Price       decimal.Decimal `bun:"price,type:decimal,notnull" json:"price"`
	FullPrice   decimal.Decimal `bun:"full_price,type:decimal,notnull" json:"full_price"`
	Currency    string          `bun:"currency,type:varchar(10),notnull" json:"currency"`
	Duration    types.Interval  `bun:"duration,type:interval,nullzero" json:"duration"`
	IsActive    bool            `bun:"is_active,type:boolean,notnull" json:"is_active"`

	MaxTotalStudents int64 `bun:"max_total_students,type:bigint,nullzero" json:"max_total_students"`
	MaxTotalProducts int64 `bun:"max_total_products,type:bigint,nullzero" json:"max_total_products"`
	MaxTotalEvents   int64 `bun:"max_total_events,type:bigint,nullzero" json:"max_total_events"`
	MaxStorageSize   int64 `bun:"max_storage_size,type:bigint,nullzero" json:"max_storage_size"`
	Personalization  bool  `bun:"personalization,type:boolean,notnull" json:"personalization"`
	TechSupport      bool  `bun:"tech_support,type:boolean,notnull" json:"tech_support"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
}
