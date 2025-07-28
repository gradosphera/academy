package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ChunkRepository struct {
	repository.Generic[model.Chunk, uuid.UUID]
}

func (r *ChunkRepository) WithTx(tx bun.Tx) *ChunkRepository {
	return &ChunkRepository{Generic: r.Generic.WithTx(tx)}
}

func NewChunkRepository(
	genericRepository repository.Generic[model.Chunk, uuid.UUID],
) *ChunkRepository {
	return &ChunkRepository{
		Generic: genericRepository,
	}
}

func (r *ChunkRepository) Create(ctx context.Context, chunk *model.Chunk) error {

	query := r.DB.NewInsert().
		On(`CONFLICT ("material_id", "index") DO NOTHING`).
		Model(chunk)

	_, err := query.
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *ChunkRepository) GetByID(
	ctx context.Context,
	materialID uuid.UUID,
	index ...int64,
) ([]*model.Chunk, error) {

	chunks := make([]*model.Chunk, 0)

	query := r.DB.NewSelect().
		Model(&chunks).
		Where(`material_id = ?`, materialID).
		Order("index")

	if len(index) != 0 {
		query = query.Where(`index IN (?)`, bun.In(index))
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return chunks, nil
}

func (r *ChunkRepository) Delete(ctx context.Context, materialID uuid.UUID) error {
	var chunk *model.Chunk

	_, err := r.DB.NewDelete().
		Model(chunk).
		Where("material_id = ?", materialID).
		Exec(ctx)

	return err
}

func (r *ChunkRepository) DeleteOlderThen(ctx context.Context, t time.Time) error {
	var chunk *model.Chunk

	_, err := r.DB.NewDelete().
		Model(chunk).
		Where("created_at < ?", t).
		Exec(ctx)

	return err
}
