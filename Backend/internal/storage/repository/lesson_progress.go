package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type LessonProgressRepository struct {
	repository.Generic[model.LessonProgress, uuid.UUID]
}

func (r *LessonProgressRepository) WithTx(tx bun.Tx) *LessonProgressRepository {
	return &LessonProgressRepository{Generic: r.Generic.WithTx(tx)}
}

func NewLessonProgressRepository(
	genericRepository repository.Generic[model.LessonProgress, uuid.UUID],
) *LessonProgressRepository {
	return &LessonProgressRepository{
		Generic: genericRepository,
	}
}

func (r *LessonProgressRepository) CreateOrUpdate(
	ctx context.Context,
	progress *model.LessonProgress,
) error {

	_, err := r.DB.NewInsert().Model(progress).
		On("CONFLICT (user_id, lesson_id) DO UPDATE").
		Set("status = ?", progress.Status).
		Set("data = ?", progress.Data).
		Set("score = ?", progress.Score).
		Set("updated_at = ?", progress.UpdatedAt).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *LessonProgressRepository) GetByID(
	ctx context.Context,
	userID uuid.UUID,
	lessonID uuid.UUID,
) (*model.LessonProgress, error) {

	progress := new(model.LessonProgress)

	query := r.DB.NewSelect().
		Model(progress).
		Where(`user_id = ?`, userID).
		Where(`lesson_id = ?`, lessonID)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return progress, nil
}

func (r *LessonProgressRepository) GetByProductID(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
) ([]*model.LessonProgress, error) {

	progress := make([]*model.LessonProgress, 0)

	query := r.DB.NewSelect().
		Model(&progress).
		Where(`user_id = ?`, userID).
		Where(`lesson_id IN (SELECT "id" FROM lessons WHERE product_id = ?)`, productID).
		Order(`lesson_id`)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return progress, nil
}

func (r *LessonProgressRepository) Find(
	ctx context.Context,
	filter *model.FilterLessonProgressRequest,
) ([]*model.LessonProgress, int, error) {

	applyFilter := func(q *bun.SelectQuery) {
		if len(filter.ProductID) != 0 {
			q = q.Join(`LEFT JOIN lessons l ON l.id = lesson_id`).
				Where(`l.product_id IN (?)`, bun.In(filter.ProductID))
		}

		if len(filter.UserID) != 0 {
			q = q.Where(`user_id IN (?)`, bun.In(filter.UserID))
		}
		if len(filter.LessonID) != 0 {
			q = q.Where(`lesson_id IN (?)`, bun.In(filter.LessonID))
		}
		if len(filter.Status) != 0 {
			q = q.Where(`status IN (?)`, bun.In(filter.Status))
		}
	}

	progress := make([]*model.LessonProgress, 0)

	countQuery := r.DB.NewSelect().
		Model(&progress)

	applyFilter(countQuery)

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*model.LessonProgress{}, total, nil
	}

	query := r.DB.NewSelect().
		Model(&progress).
		Order(`updated_at DESC`).
		Limit(int(filter.Limit))

	applyFilter(query)

	if filter.Offset != 0 {
		query = query.Offset(int(filter.Offset))
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, total, err
	}

	return progress, total, nil
}

func (r *LessonProgressRepository) Update(ctx context.Context, progress *model.LessonProgress) error {
	_, err := r.DB.NewUpdate().
		Model(progress).
		WherePK().
		Exec(ctx)

	return err
}

func (r *LessonProgressRepository) Delete(
	ctx context.Context, userID uuid.UUID, productIDs []uuid.UUID,
) ([]*model.LessonProgress, error) {

	progress := make([]*model.LessonProgress, 0)

	_, err := r.DB.NewDelete().
		Model(&progress).
		Where(`user_id = ?`, userID).
		Where(`lesson_id IN (SELECT id FROM lessons WHERE product_id IN (?))`, bun.In(productIDs)).
		Returning(`*`).
		Exec(ctx)

	if err != nil {
		return nil, err

	}

	return progress, nil
}
