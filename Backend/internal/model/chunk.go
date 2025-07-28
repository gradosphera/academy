package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Chunk struct {
	bun.BaseModel `bun:"table:chunks"`

	MaterialID uuid.UUID `bun:"material_id,pk,type:uuid,notnull" json:"material_id"`
	MiniAppID  uuid.UUID `bun:"mini_app_id,type:uuid,notnull" json:"mini_app_id"`
	Index      int64     `bun:"index,pk,type:int,notnull" json:"index"`
	Size       int64     `bun:"size,type:int,notnull,default:0" json:"size"`
	CreatedAt  time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
}

func NewChunk(miniAppID, materialID uuid.UUID, index int64, size int64) *Chunk {
	now := time.Now().UTC()
	return &Chunk{
		MaterialID: materialID,
		MiniAppID:  miniAppID,
		Index:      index,
		Size:       size,
		CreatedAt:  now,
	}
}

type CreateChunkRequest struct {
	Hashsum string `json:"hashsum"`
}
