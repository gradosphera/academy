package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProductAccess struct {
	bun.BaseModel `bun:"table:product_access"`

	UserID    uuid.UUID `bun:"user_id,pk,type:uuid,notnull" json:"user_id"`
	ProductID uuid.UUID `bun:"product_id,pk,type:uuid,notnull" json:"product_id"`

	DeletedReason string `bun:"deleted_reason,type:varchar(255),notnull" json:"deleted_reason,omitempty"`

	DeletedAt *time.Time `bun:"deleted_at,type:timestamptz,nullzero" json:"deleted_at,omitempty"`
	UpdatedAt time.Time  `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time  `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
}

func NewProductAccess(userID, productID uuid.UUID) *ProductAccess {
	now := time.Now().UTC()
	return &ProductAccess{
		UserID:    userID,
		ProductID: productID,
		UpdatedAt: now,
		CreatedAt: now,
	}
}
