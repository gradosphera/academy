package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ReviewRepository struct {
	repository.Generic[model.Review, uuid.UUID]
}

func (r *ReviewRepository) WithTx(tx bun.Tx) *ReviewRepository {
	return &ReviewRepository{Generic: r.Generic.WithTx(tx)}
}

func NewReviewRepository(
	genericRepository repository.Generic[model.Review, uuid.UUID],
) *ReviewRepository {
	return &ReviewRepository{
		Generic: genericRepository,
	}
}

func (r *ReviewRepository) Create(ctx context.Context, review *model.Review) error {
	_, err := r.DB.NewInsert().Model(review).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	review := new(model.Review)

	query := r.DB.NewSelect().
		Model(review).
		Where(`id = ?`, id)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return review, nil
}

func (r *ReviewRepository) GetByUser(
	ctx context.Context,
	userID, productID uuid.UUID,
) ([]*model.Review, error) {

	reviews := make([]*model.Review, 0)

	query := r.DB.NewSelect().
		Model(&reviews).
		Join(`JOIN lessons AS l ON l.id = lesson_id AND l.product_id = ?`, productID).
		Where(`user_id = ?`, userID)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return reviews, nil
}

func (r *ReviewRepository) Update(ctx context.Context, review *model.Review) error {
	_, err := r.DB.NewUpdate().
		Model(review).
		WherePK().
		Exec(ctx)

	return err
}

func (r *ReviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var review *model.Review

	_, err := r.DB.NewDelete().
		Model(review).
		Where("id = ?", id).
		Exec(ctx)

	return err
}
