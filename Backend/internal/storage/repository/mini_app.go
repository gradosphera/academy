package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var ErrDuplicatedKey = errors.New("duplicated key")

type MiniAppRepository struct {
	repository.Generic[model.MiniApp, uuid.UUID]
}

func (r *MiniAppRepository) WithTx(tx bun.Tx) *MiniAppRepository {
	return &MiniAppRepository{Generic: r.Generic.WithTx(tx)}
}

func NewMiniAppRepository(
	genericRepository repository.Generic[model.MiniApp, uuid.UUID],
) *MiniAppRepository {
	return &MiniAppRepository{
		Generic: genericRepository,
	}
}

func (r *MiniAppRepository) Create(ctx context.Context, miniApp *model.MiniApp) error {
	_, err := r.DB.NewInsert().Model(miniApp).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *MiniAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MiniApp, error) {
	miniApp := new(model.MiniApp)

	query := r.DB.NewSelect().
		Model(miniApp).
		Where(`"mini_app".id = ?`, id).
		Relation("Owner").
		Relation("Products", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("index")
		}).
		Relation("Products.Levels", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("index", "price")
		}).
		Relation("Slides", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("index").Where("category = ?", model.MaterialCategorySlides)
		}).
		Relation("TOS", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("index").Where("category = ?", model.MaterialCategoryTOS)
		}).
		Relation("PrivacyPolicy", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("index").Where("category = ?", model.MaterialCategoryPrivacyPolicy)
		})

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return miniApp, nil
}

func (r *MiniAppRepository) GetPlanByPlanID(
	ctx context.Context, planID model.PlanID,
) (*model.Plan, error) {

	plan := new(model.Plan)

	query := r.DB.NewSelect().
		Model(plan).
		Where("id = ?", planID)

	err := query.Scan(ctx, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (r *MiniAppRepository) GetByName(ctx context.Context, name string) (*model.MiniApp, error) {
	miniApp := new(model.MiniApp)

	query := r.DB.NewSelect().
		Model(miniApp).
		Where("name = ?", name).
		Relation("Owner")

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return miniApp, nil
}

func (r *MiniAppRepository) GetByOwnerTelegramID(
	ctx context.Context,
	ownerTelegramID int64,
) (*model.MiniApp, error) {

	var miniApp = new(model.MiniApp)

	err := r.DB.NewSelect().
		Model(miniApp).
		Where("owner_telegram_id = ?", ownerTelegramID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return miniApp, nil
}

func (r *MiniAppRepository) GetByModTelegramID(
	ctx context.Context, modTelegramID int64,
) ([]*model.MiniApp, error) {

	miniApps := make([]*model.MiniApp, 0)

	query := r.DB.NewRaw(`
	SELECT * FROM mini_apps
	WHERE id IN (
		SELECT DISTINCT u.mini_app_id FROM users AS u
		JOIN mod_invites AS mi ON mi.user_id = u.id
		WHERE u.telegram_id = ?
	)
	`, modTelegramID)

	err := query.Scan(ctx, &miniApps)
	if err != nil {
		return nil, err
	}

	return miniApps, nil
}

func (r *MiniAppRepository) Update(ctx context.Context, model *model.MiniApp) error {
	query := r.DB.NewUpdate().
		Model(model).
		WherePK()

	_, err := query.Exec(ctx)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("%w: %w", ErrDuplicatedKey, err)
		}
		return err
	}

	return nil
}

func (r *MiniAppRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*model.MiniApp)(nil)).
		Set(`deleted_at = now()`).
		Where(`id = ?`, id).
		Exec(ctx)

	return err
}

func (r *MiniAppRepository) Restore(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*model.MiniApp)(nil)).
		Set(`deleted_at = NULL`).
		Where(`id = ?`, id).
		Exec(ctx)

	return err
}

func (r *MiniAppRepository) DeleteOld(ctx context.Context, olderThan time.Time) ([]uuid.UUID, error) {
	miniApps := make([]*model.MiniApp, 0)

	query := r.DB.NewDelete().
		Model(&miniApps).
		Where(`deleted_at < ?`, olderThan).
		Returning(`id`)

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return []uuid.UUID{}, nil
	}

	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(miniApps))
	for i, miniApp := range miniApps {
		ids[i] = miniApp.ID
	}

	return ids, nil
}

func (r *MiniAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.DB.NewDelete().
		Model((*model.MiniApp)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	return err
}

func (r *MiniAppRepository) Analytics(
	ctx context.Context,
	miniAppID uuid.UUID,
	req *model.AnalyticsRequest,
) (*model.Analytics, error) {

	var analytics model.Analytics

	periodStart := time.Now().AddDate(
		-int(req.TimePeriod.Months/12),
		-int(req.TimePeriod.Months%12),
		-int(req.TimePeriod.Days),
	).Add(-time.Duration(req.TimePeriod.Microseconds) * time.Microsecond)

	// TODO: This query should not include payments for plans.

	err := r.DB.NewRaw(`
	WITH mini_app_students AS (
		SELECT * FROM users
		WHERE mini_app_id = ? AND "role" = 'student'
	), mini_app_payments AS (
		SELECT * FROM payments
		WHERE mini_app_id = ? AND status = 'completed'
	)
	SELECT
		(
			SELECT COUNT(*) FROM mini_app_students WHERE ? < created_at
		) AS new_students,
		(
			SELECT COUNT(*) FROM mini_app_students
		) AS total_students,
		(
			SELECT COALESCE(SUM("amount_usd"), 0) FROM mini_app_payments
			WHERE ? < updated_at
		) AS money_earned,
		(
			SELECT COALESCE(SUM("amount_usd"), 0) FROM mini_app_payments
		) AS total_money_earned
	`, miniAppID, miniAppID, periodStart, periodStart).
		Scan(ctx, &analytics)

	if err != nil {
		return nil, err
	}

	var productsAnalytics []*model.ProductAnalytics

	err = r.DB.NewRaw(`
	WITH mini_app_products AS (
		SELECT products.*
		FROM products
		WHERE mini_app_id = ?
	)
	SELECT
		product_id,
		SUM(new_students) AS new_students,
		SUM(total_students) AS total_students,
		SUM(money_earned) AS money_earned,
		SUM(total_money_earned) AS total_money_earned
	FROM (
		SELECT
			p.id AS product_id,
			p.index,
			p.created_at,
			COUNT( CASE WHEN ? < pa.updated_at THEN students.id END ) AS new_students,
			COUNT( students.id ) AS total_students,
			0 AS money_earned,
			0 AS total_money_earned
		FROM mini_app_products AS p
		LEFT JOIN product_access AS pa ON pa.product_id = p.id AND pa.deleted_at IS NULL
		LEFT JOIN users AS students ON students.id = pa.user_id AND students.role = 'student'
		GROUP BY p.id, p.index, p.created_at

		UNION ALL

		SELECT
			p.id AS product_id,
			p.index,
			p.created_at,
			0 AS new_students,
			0 AS total_students,
			COALESCE(
				SUM(
					CASE WHEN ? < payments.updated_at THEN COALESCE(payments.amount_usd, 0) END
				), 0
			) AS money_earned,
			COALESCE(
				SUM(
					COALESCE(payments.amount_usd, 0)
				), 0
			) AS total_money_earned
		FROM mini_app_products AS p
		LEFT JOIN payments ON payments.product_id = p.id AND payments.status = 'completed'
		GROUP BY p.id, p.index, p.created_at
	) t1
	GROUP BY product_id, index, created_at
	ORDER BY index, created_at
	`, miniAppID, periodStart, periodStart).
		Scan(ctx, &productsAnalytics)

	if err != nil {
		return nil, err
	}

	analytics.Products = productsAnalytics

	return &analytics, nil
}

func (r *MiniAppRepository) GetInfo(ctx context.Context, miniAppID uuid.UUID) (*model.MiniAppInfo, error) {
	var info model.MiniAppInfo

	err := r.DB.NewRaw(`
	SELECT
		storage_size,
		total_products,
		total_students,
		total_events,
		max_storage_size,
		max_total_products,
		max_total_students,
		max_total_events
	FROM mini_apps
	WHERE id = ?
	`, miniAppID).
		Scan(ctx, &info)

	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (r *MiniAppRepository) CreateModInvite(
	ctx context.Context, invite *model.ModInvite,
) error {

	_, err := r.DB.NewInsert().
		Model(invite).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *MiniAppRepository) CreateModInvitePermission(
	ctx context.Context, invitePermission *model.ModInvitePermission,
) error {

	_, err := r.DB.NewInsert().
		Model(invitePermission).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *MiniAppRepository) DeleteModInvitePermissions(
	ctx context.Context, inviteID uuid.UUID,
) error {

	_, err := r.DB.NewDelete().
		Model((*model.ModInvitePermission)(nil)).
		Where(`invite_id = ?`, inviteID).
		Exec(ctx)

	if err != nil {
		return err
	}

	return err
}

func (r *MiniAppRepository) GetModInvite(
	ctx context.Context,
	id uuid.UUID,
) (*model.ModInvite, error) {

	invite := new(model.ModInvite)

	query := r.DB.NewSelect().
		Model(invite).
		Where(`id = ?`, id)

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return invite, nil
}

func (r *MiniAppRepository) CheckPermission(
	ctx context.Context,
	userID uuid.UUID,
	permissionName ...model.PermissionName,
) (bool, error) {

	permissions := make([]*model.ModInvitePermission, 0)

	query := r.DB.NewRaw(`
	SELECT mip.* FROM mod_invites AS mi
	JOIN mod_invite_permissions AS mip ON mip.invite_id = mi.id
		AND mip.permission_name IN (?)
	WHERE "user_id" = ?
	`, bun.In(permissionName), userID)

	err := query.Scan(ctx, &permissions)
	if err != nil {
		return false, err
	}

	return 0 < len(permissions), nil
}

func (r *MiniAppRepository) GetPermissions(
	ctx context.Context,
	userID uuid.UUID,
) ([]*model.Permission, error) {

	permissions := make([]*model.Permission, 0)

	query := r.DB.NewRaw(`
	SELECT
		p.name,
		p.description
	FROM mod_invites AS mi
	JOIN mod_invite_permissions AS mip ON mip.invite_id = mi.id
	JOIN permissions AS p ON p.name = mip.permission_name
	WHERE "user_id" = ?
	GROUP BY p.name, p.description
	`, userID)

	err := query.Scan(ctx, &permissions)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *MiniAppRepository) ListPermissions(ctx context.Context) ([]*model.Permission, error) {
	permissions := make([]*model.Permission, 0)

	err := r.DB.NewSelect().Model(&permissions).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *MiniAppRepository) UpdateModInvite(
	ctx context.Context,
	invite *model.ModInvite,
) error {

	_, err := r.DB.NewUpdate().
		Model(invite).
		Where(`id = ?`, invite.ID).
		Where(`user_id IS NULL`).
		Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (r *MiniAppRepository) ModInvites(
	ctx context.Context,
	miniAppID uuid.UUID,
	filter *model.FilterModInvitesRequest,
) ([]*model.ModInvite, int, error) {

	invites := make([]*model.ModInvite, 0)

	countQuery := r.DB.NewSelect().
		Model(&invites)

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*model.ModInvite{}, total, nil
	}

	query := r.DB.NewSelect().
		Model(&invites).
		Relation("User").
		Relation("Permissions").
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

func (r *MiniAppRepository) DeleteModInvite(
	ctx context.Context,
	miniAppID, inviteID uuid.UUID,
) error {

	_, err := r.DB.NewDelete().
		Model((*model.ModInvite)(nil)).
		Where("mini_app_id = ?", miniAppID).
		Where("id = ?", inviteID).
		Exec(ctx)

	if err != nil {
		return err
	}

	return err
}

func (r *MiniAppRepository) FindTonAddresses(ctx context.Context) ([]string, error) {
	query := r.DB.NewRaw(`
	SELECT DISTINCT payment_metadata ->> 'ton_address' AS address FROM mini_apps
	WHERE payment_metadata ->> 'ton_address' IS NOT NULL
	AND payment_metadata ->> 'ton_address' <> ''
	`)

	var addresses []string
	if err := query.Scan(ctx, &addresses); err != nil {
		return nil, err
	}

	return addresses, nil
}
