package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserRepository struct {
	repository.Generic[model.User, uuid.UUID]
}

func (r *UserRepository) WithTx(tx bun.Tx) *UserRepository {
	return &UserRepository{Generic: r.Generic.WithTx(tx)}
}

func NewUserRepository(
	genericRepository repository.Generic[model.User, uuid.UUID],
) *UserRepository {

	return &UserRepository{
		Generic: genericRepository,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	_, err := r.DB.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := new(model.User)
	err := r.DB.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByTelegramID(
	ctx context.Context,
	telegramID int64,
	miniAppID uuid.UUID,
) (*model.User, error) {

	var user = new(model.User)

	err := r.DB.NewSelect().
		Model(user).
		Where("telegram_id = ?", telegramID).
		Where("mini_app_id = ?", miniAppID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	_, err := r.DB.NewUpdate().
		Model(user).
		WherePK().
		ExcludeColumn("created_at").
		Exec(ctx)

	return err
}

func (r *UserRepository) StudentStats(
	ctx context.Context,
	miniAppID, userID uuid.UUID,
) ([]*model.StudentProductStats, error) {

	stats := make([]*model.StudentStats, 0)

	err := r.DB.NewRaw(`
	SELECT
		p.id AS product_id,
		p.title AS product_title,
		p.cover AS product_cover,
		p.content_type AS product_content_type,
		pa.created_at AS joined_at,
		pa.deleted_reason AS product_access_deleted_reason,
		pa.deleted_at AS product_access_deleted_at,
		l.id AS lesson_id,
		l.module_name,
		l.content_type AS lesson_content_type,
		l.title AS lesson_title,
		COALESCE(lp.status::TEXT, '') AS progress_status,
		COALESCE(lp.score, 0) AS score,
		COALESCE(r."text", '') AS review_text,
		COALESCE(r.score, 0) AS review_score
	FROM lessons AS l
	JOIN product_access AS pa ON pa.product_id = l.product_id
		AND pa.user_id = ?
	JOIN products AS p ON p.id = l.product_id
		AND p.mini_app_id = ?
		AND p.is_active = TRUE
	LEFT JOIN lesson_progress AS lp ON lp.lesson_id = l.id
		AND lp.user_id = pa.user_id
	LEFT JOIN reviews AS r ON r.lesson_id = lp.lesson_id
		AND r.user_id = lp.user_id 
	ORDER BY p.index, p.id, l.index, l.id
	`, userID, miniAppID).
		Scan(ctx, &stats)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	result := make([]*model.StudentProductStats, 0)

	prevProductID := uuid.Nil
	for _, st := range stats {
		if st.ProductID != prevProductID {
			prevProductID = st.ProductID

			result = append(result, &model.StudentProductStats{
				ProductID:   st.ProductID,
				Title:       st.ProductTitle,
				Cover:       st.ProductCover,
				ContentType: st.ProductContentType,
				JoinedAt:    st.JoinedAt,

				ProductAccessDeletedReason: st.ProductAccessDeletedReason,
				ProductAccessDeletedAt:     st.ProductAccessDeletedAt,

				LessonStats: []*model.StudentLessonStats{},
			})
		}

		result[len(result)-1].LessonStats = append(result[len(result)-1].LessonStats,
			&model.StudentLessonStats{
				LessonID:    st.LessonID,
				ModuleName:  st.ModuleName,
				ContentType: st.LessonContentType,
				Title:       st.LessonTitle,

				ProgressStatus: st.ProgressStatus,
				Score:          st.Score,
				ReviewScore:    st.ReviewScore,
				ReviewText:     st.ReviewText,
			})
	}

	return result, nil
}

func (r *UserRepository) GetByAvatarFilename(
	ctx context.Context,
	miniAddID uuid.UUID,
	filename string,
) (*model.User, error) {

	user := new(model.User)

	err := r.DB.NewSelect().
		Model(user).
		Where("mini_app_id = ?", miniAddID).
		Where("avatar = ?", filename).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) RestrictProductAccess(
	ctx context.Context,
	userID uuid.UUID, req *model.BanUserRequest,
) error {

	if len(req.ProductID) == 0 {
		return nil
	}

	accesses := make([]*model.ProductAccess, 0)
	for _, productID := range req.ProductID {
		accesses = append(accesses, model.NewProductAccess(userID, productID))
	}

	_, err := r.DB.NewInsert().
		Model(&accesses).
		On(`CONFLICT (user_id, product_id) DO NOTHING`).
		Exec(ctx)

	if err != nil {
		return err
	}

	now := time.Now()

	query := r.DB.NewUpdate().
		Model((*model.ProductAccess)(nil)).
		Set(`deleted_reason = ?`, req.Reason).
		Set(`deleted_at = ?`, now).
		Set(`updated_at = ?`, now).
		Where(`product_id IN (?)`, bun.In(req.ProductID)).
		Where(`user_id = ?`, userID)

	_, err = query.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) AllowProductAccess(
	ctx context.Context,
	userID uuid.UUID, req *model.UnbanUserRequest,
) error {

	if len(req.ProductID) == 0 {
		return nil
	}

	now := time.Now()

	query := r.DB.NewUpdate().
		Model((*model.ProductAccess)(nil)).
		Set(`deleted_reason = ''`).
		Set(`deleted_at = NULL`).
		Set(`updated_at = ?`, now).
		Where(`product_id IN (?)`, bun.In(req.ProductID)).
		Where(`user_id = ?`, userID)

	_, err := query.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) ListRestrictedUsers(
	ctx context.Context,
	miniAppID uuid.UUID,
	filter *model.ListBannedUserRequest,
) ([]*model.User, int, error) {

	users := make([]*model.User, 0)

	countQuery := r.DB.NewSelect().
		Model(&users).
		Where(`id IN (SELECT DISTINCT user_id FROM product_access WHERE deleted_at IS NOT NULL)`)

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*model.User{}, total, nil
	}

	query := r.DB.NewSelect().
		Model(&users).
		Where(`id IN (SELECT DISTINCT user_id FROM product_access WHERE deleted_at IS NOT NULL)`).
		Relation("ProductAccess", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("deleted_at IS NOT NULL")
		}).
		Order(`created_at DESC`).
		Limit(int(filter.Limit))

	if filter.Offset != 0 {
		query = query.Offset(int(filter.Offset))
	}

	err = query.Scan(ctx)
	if err != nil {
		return nil, total, err
	}

	return users, total, nil
}

func (r *UserRepository) Levels(
	ctx context.Context,
	miniAppID, userID, productID uuid.UUID,
) ([]*model.StudentProductLevels, error) {

	levels := make([]*model.StudentProductLevels, 0)

	err := r.DB.NewRaw(`
	SELECT
		DISTINCT ON (payments.updated_at) payments.updated_at AS payment_updated_at,
		payments.id AS payment_id,
		pl.id AS product_level_id,
		pl.name AS product_level_name,
		pl.description AS product_level_description,
		payments.amount AS paid_price,
		payments.currency AS paid_currency,
		payments.access_start + payments.access_duration AS ends_at
	FROM payments
	JOIN product_levels AS pl ON pl.id = payments.product_level_id AND pl.product_id = ?
	WHERE payments.mini_app_id = ? AND payments.user_id = ? AND payments.status = ?
	ORDER BY payments.updated_at DESC
	`, productID, miniAppID, userID, model.PaymentStatusCompleted).
		Scan(ctx, &levels)

	if err != nil {
		return nil, err
	}

	return levels, nil
}
