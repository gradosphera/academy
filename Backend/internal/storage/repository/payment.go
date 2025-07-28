package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/types"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PaymentRepository struct {
	repository.Generic[model.Payment, uuid.UUID]
}

func (r *PaymentRepository) WithTx(tx bun.Tx) *PaymentRepository {
	return &PaymentRepository{Generic: r.Generic.WithTx(tx)}
}

func NewPaymentRepository(
	genericRepository repository.Generic[model.Payment, uuid.UUID],
) *PaymentRepository {
	return &PaymentRepository{
		Generic: genericRepository,
	}
}

func (r *PaymentRepository) Create(ctx context.Context, payment ...*model.Payment) error {
	_, err := r.DB.NewInsert().Model(&payment).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	payment := new(model.Payment)

	query := r.DB.NewSelect().
		Model(payment).
		Relation("User").
		Relation("MiniApp").
		Where(`payment.id = ?`, id)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *PaymentRepository) Find(
	ctx context.Context,
	filter *model.FilterPayments,
) ([]*model.Payment, int, error) {

	applyFilter := func(q *bun.SelectQuery) {
		q = q.Where(`payment.mini_app_id = ?`, filter.MiniAppID)

		q = q.Where(`payment.currency <> ''`) // Ignore payments created by NewFreePaymentForProductLevel.

		if len(filter.ID) != 0 {
			q = q.Where("id IN (?)", bun.In(filter.ID))
		}
		if len(filter.UserID) != 0 {
			q = q.Where("user_id IN (?)", bun.In(filter.UserID))
		}
		if len(filter.Status) != 0 {
			q = q.Where("status IN (?)", bun.In(filter.Status))
		}
	}

	payment := make([]*model.Payment, 0)

	countQuery := r.DB.NewSelect().
		Model(&payment)

	applyFilter(countQuery)

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*model.Payment{}, total, nil
	}

	query := r.DB.NewSelect().
		Model(&payment).
		Relation("User").
		Relation("ProductLevel").
		Relation("Plan").
		Order("created_at DESC").
		Limit(int(filter.Limit))

	applyFilter(query)

	if filter.Offset != 0 {
		query = query.Offset(int(filter.Offset))
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, total, err
	}

	return payment, total, nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	_, err := r.DB.NewUpdate().
		Model(payment).
		WherePK().
		Exec(ctx)

	return err
}

func (r *PaymentRepository) UpdateAccessStart(
	ctx context.Context,
	productID uuid.UUID,
	newAccessStart types.Time,
) error {

	query := r.DB.NewUpdate().
		Model((*model.Payment)(nil)).
		Set(`access_start = ?`, newAccessStart).
		Where(`product_level_id IN (SELECT id FROM product_levels WHERE product_id = ?)`, productID).
		Where(`CURRENT_TIMESTAMP < access_start`)

	_, err := query.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var payment *model.Payment

	_, err := r.DB.NewDelete().
		Model(payment).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *PaymentRepository) StudentsPayments(
	ctx context.Context,
	miniAppID uuid.UUID,
	dateFrom, dateTo time.Time,
) ([]*model.StudentsPayment, error) {

	result := make([]*model.StudentsPayment, 0)

	err := r.DB.NewRaw(`
	SELECT
		payments.id AS payment_id,
		u.id AS user_id,
		u.telegram_id,
		u.telegram_username,
		COALESCE(p.title, '') AS product_name,
		COALESCE(pl."name", '') AS product_level_name,
		payments.amount_usd,
		payments.updated_at AS paid_at
	FROM payments
	JOIN users AS u ON u.id = payments.user_id AND u.role = 'student'
	LEFT JOIN products AS p ON p.id = payments.product_id
	LEFT JOIN product_levels AS pl ON pl.id = payments.product_level_id 
	WHERE payments.mini_app_id = ?
		AND payments.status = 'completed'
		AND payments.updated_at BETWEEN ? AND ?
	ORDER BY payments.updated_at
	`, miniAppID, dateFrom, dateTo.AddDate(0, 0, 1)).
		Scan(ctx, &result)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}
