package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/types"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProductLevelRepository struct {
	repository.Generic[model.ProductLevel, uuid.UUID]
}

func (r *ProductLevelRepository) WithTx(tx bun.Tx) *ProductLevelRepository {
	return &ProductLevelRepository{Generic: r.Generic.WithTx(tx)}
}

func NewProductLevelRepository(
	genericRepository repository.Generic[model.ProductLevel, uuid.UUID],
) *ProductLevelRepository {
	return &ProductLevelRepository{
		Generic: genericRepository,
	}
}

func (r *ProductLevelRepository) Create(
	ctx context.Context,
	productLevel *model.ProductLevel,
) error {

	_, err := r.DB.NewInsert().Model(productLevel).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductLevelRepository) Update(ctx context.Context, model *model.ProductLevel) error {
	_, err := r.DB.NewUpdate().
		Model(model).
		WherePK().
		Exec(ctx)

	return err
}

func (r *ProductLevelRepository) UpdateDuration(
	ctx context.Context,
	productID uuid.UUID,
	newDuration types.Interval,
) error {

	query := r.DB.NewUpdate().
		Model((*model.ProductLevel)(nil)).
		Set(`duration = ?`, newDuration).
		Where(`product_id = ?`, productID)

	_, err := query.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductLevelRepository) GetByID(ctx context.Context, id ...uuid.UUID) ([]*model.ProductLevel, error) {
	productLevels := make([]*model.ProductLevel, 0)

	err := r.DB.NewSelect().
		Model(&productLevels).
		Where("id IN (?)", bun.In(id)).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return productLevels, nil
}

func (r *ProductLevelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var productLevel *model.ProductLevel

	_, err := r.DB.NewDelete().
		Model(productLevel).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *ProductLevelRepository) AddLessonRelations(
	ctx context.Context,
	productLevelID uuid.UUID, lessonsIDs ...uuid.UUID,
) error {

	if len(lessonsIDs) == 0 {
		return nil
	}

	productLevelLessons := make([]*model.ProductLevelLesson, 0, len(lessonsIDs))
	for _, lessonID := range lessonsIDs {
		productLevelLessons = append(productLevelLessons,
			model.NewProductLevelLesson(productLevelID, lessonID))
	}

	_, err := r.DB.NewInsert().
		Model(&productLevelLessons).
		On("CONFLICT (product_level_id, lesson_id) DO NOTHING").
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *ProductLevelRepository) RemoveLessonRelations(
	ctx context.Context,
	productLevelID uuid.UUID, lessonsIDs ...uuid.UUID,
) error {

	if len(lessonsIDs) == 0 {
		return nil
	}

	var productLevelLesson *model.ProductLevelLesson

	query := r.DB.NewDelete().
		Model(productLevelLesson).
		Where("product_level_id = ?", productLevelID).
		Where("lesson_id IN (?)", bun.In(lessonsIDs))

	_, err := query.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductLevelRepository) IsProductLevelUnlocked(
	ctx context.Context,
	productLevelID, userID uuid.UUID,
) (bool, error) {

	payments := make([]*model.Payment, 0)

	now := time.Now()

	err := r.DB.NewRaw(`
	SELECT payments.* FROM payments
	WHERE payments.product_level_id = ?
		AND payments.user_id = ?
		AND payments.status = ?
		AND (payments.access_duration IS NULL OR ? < (payments.access_start + payments.access_duration))
	LIMIT 1
	`, now, productLevelID, userID, model.PaymentStatusCompleted).
		Scan(ctx, &payments)

	if err != nil {
		return false, err
	}

	return len(payments) != 0, nil
}

func (r *ProductLevelRepository) GetProducts(ctx context.Context, id []uuid.UUID) ([]*model.Product, error) {
	products := make([]*model.Product, 0)

	err := r.DB.NewRaw(`
	SELECT p.* FROM products AS p
	JOIN product_levels AS pl ON pl.product_id = p.id AND pl.id IN (?)
	`, bun.In(id)).
		Scan(ctx, &products)

	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductLevelRepository) CreateProductLevelInvite(ctx context.Context, invite *model.ProductLevelInvite) error {
	_, err := r.DB.NewInsert().Model(invite).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductLevelRepository) FindProductLevelInvites(
	ctx context.Context,
	inviteID uuid.UUID,
	userID uuid.UUID,
	productID uuid.UUID,
) ([]*model.ProductLevelInvite, error) {

	invites := make([]*model.ProductLevelInvite, 0)

	query := r.DB.NewSelect().Model(&invites)

	if inviteID != uuid.Nil {
		query = query.Where(`id = ?`, inviteID)
	}
	if userID != uuid.Nil {
		query = query.Where(`user_id = ?`, userID)
	}
	if productID != uuid.Nil {
		query = query.Where(`product_level_id IN (SELECT id FROM product_levels WHERE product_id = ?)`,
			productID)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, err
	}

	return invites, nil
}

func (r *ProductLevelRepository) UpdateInvite(
	ctx context.Context,
	invite *model.ProductLevelInvite,
) error {

	res, err := r.DB.NewUpdate().
		Model(invite).
		Where(`id = ?`, invite.ID).
		Where(`user_id IS NULL`).
		Exec(ctx)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("no invites was claimed")
	}

	return nil
}

func (r *ProductLevelRepository) ProductLevelInvites(
	ctx context.Context,
	productID uuid.UUID,
	filter *model.FilterProductLevelInvitesRequest,
) ([]*model.ProductLevelInvite, int, error) {

	invites := make([]*model.ProductLevelInvite, 0)

	countQuery := r.DB.NewSelect().
		Model(&invites)

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*model.ProductLevelInvite{}, total, nil
	}

	query := r.DB.NewSelect().
		Model(&invites).
		Relation("User").
		Order(`created_at DESC`).
		Limit(int(filter.Limit))

	if filter.Offset != 0 {
		query = query.Offset(int(filter.Offset))
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, total, err
	}

	return invites, total, nil
}
