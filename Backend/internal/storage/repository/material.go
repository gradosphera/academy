package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MaterialRepository struct {
	repository.Generic[model.Material, uuid.UUID]
}

func (r *MaterialRepository) WithTx(tx bun.Tx) *MaterialRepository {
	return &MaterialRepository{Generic: r.Generic.WithTx(tx)}
}

func NewMaterialRepository(
	genericRepository repository.Generic[model.Material, uuid.UUID],
) *MaterialRepository {
	return &MaterialRepository{
		Generic: genericRepository,
	}
}

func (r *MaterialRepository) Create(ctx context.Context, material *model.Material) error {
	_, err := r.DB.NewInsert().Model(material).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *MaterialRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Material, error) {
	material := new(model.Material)

	query := r.DB.NewSelect().
		Model(material).
		Where(`id = ?`, id)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return material, nil
}

func (r *MaterialRepository) GetByFilename(ctx context.Context, filename string) (*model.Material, error) {
	material := new(model.Material)

	query := r.DB.NewSelect().
		Model(material).
		Where(`filename = ?`, filename)

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return material, nil
}

func (r *MaterialRepository) FindByStatus(
	ctx context.Context, status model.MaterialStatus, withMetadata bool,
	limit, offset int,
) ([]*model.Material, error) {

	materials := make([]*model.Material, 0)

	query := r.DB.NewSelect().
		Model(&materials).
		Where(`status = ?`, status).
		Order("updated_at")

	if withMetadata {
		query = query.Where(`metadata IS NOT NULL`)
	} else {
		query = query.Where(`metadata IS NULL`)
	}

	if limit != 0 {
		query = query.Limit(limit)
	}
	if offset != 0 {
		query = query.Offset(offset)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return materials, nil
}

func (r *MaterialRepository) Update(ctx context.Context, material *model.Material) error {
	_, err := r.DB.NewUpdate().
		Model(material).
		WherePK().
		Exec(ctx)

	return err
}

func (r *MaterialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var material *model.Material

	_, err := r.DB.NewDelete().
		Model(material).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *MaterialRepository) FindMuxAssetsToDelete(ctx context.Context, limit int) ([]string, error) {
	assets := make([]string, 0)

	query := r.DB.NewSelect().
		ColumnExpr(`asset_id`).
		TableExpr(`mux_assets_to_delete`).
		Order("created_at")

	if limit != 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(ctx, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

func (r *MaterialRepository) FullyDeleleMuxAsset(ctx context.Context, assetID string) error {
	_, err := r.DB.NewDelete().
		TableExpr(`mux_assets_to_delete`).
		Where("asset_id = ?", assetID).
		Exec(ctx)

	return err
}
