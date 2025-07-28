package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProductLevelInvite struct {
	bun.BaseModel `bun:"table:product_level_invites"`

	ID             uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	UserID         uuid.UUID `bun:"user_id,type:uuid,nullzero" json:"user_id"`
	ProductLevelID uuid.UUID `bun:"product_level_id,type:uuid,notnull" json:"product_level_id"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	User *User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

func NewProductLevelInvite(productLevelID uuid.UUID) *ProductLevelInvite {
	now := time.Now().UTC()
	return &ProductLevelInvite{
		ID:             uuid.New(),
		ProductLevelID: productLevelID,
		UpdatedAt:      now,
		CreatedAt:      now,
	}
}

type FilterProductLevelInvitesRequest struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}
