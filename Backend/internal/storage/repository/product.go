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

type ProductRepository struct {
	repository.Generic[model.Product, uuid.UUID]
}

func (r *ProductRepository) WithTx(tx bun.Tx) *ProductRepository {
	return &ProductRepository{Generic: r.Generic.WithTx(tx)}
}

func NewProductRepository(
	genericRepository repository.Generic[model.Product, uuid.UUID],
) *ProductRepository {
	return &ProductRepository{
		Generic: genericRepository,
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	_, err := r.DB.NewInsert().Model(product).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
	includeRelations bool,
) (*model.Product, error) {

	product := new(model.Product)

	query := r.DB.NewSelect().
		Model(product).
		Where(`id = ?`, id)

	if includeRelations {
		query = query.Relation("Lessons", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("index")
		}).
			Relation("Lessons.Materials", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Order("category", "index")
			}).
			Relation("Levels", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Order("index", "price")
			}).
			Relation("Levels.ProductLevelLessons").
			Relation("Levels.Bonus", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Order("index")
			})
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *ProductRepository) Update(
	ctx context.Context,
	model *model.Product,
) error {

	_, err := r.DB.NewUpdate().
		Model(model).
		WherePK().
		Exec(ctx)

	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var model *model.Product

	_, err := r.DB.NewDelete().
		Model(model).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *ProductRepository) GetProductAccess(
	ctx context.Context,
	userID, productID uuid.UUID,
) (*model.ProductAccess, error) {

	var productAccess *model.ProductAccess

	query := r.DB.NewSelect().
		Model(productAccess).
		Where(`user_id = ?`, userID).
		Where(`product_id = ?`, productID)

	err := query.Scan(ctx)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return productAccess, nil
}

// CheckProductAccess same as just getting but also creates new if not exists and
// sets updated_at.
func (r *ProductRepository) CheckProductAccess(
	ctx context.Context,
	productAccess *model.ProductAccess,
) (*model.ProductAccess, error) {

	_, err := r.DB.NewInsert().
		Model(productAccess).
		On(`CONFLICT (user_id, product_id) DO UPDATE`).
		Set(`updated_at = ?`, time.Now()).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	err = r.DB.NewSelect().
		Model(productAccess).
		WherePK().
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return productAccess, nil
}

func (r *ProductRepository) ProductAccessByUser(
	ctx context.Context,
	miniAppID, userID uuid.UUID,
) ([]*model.ProductAccess, error) {

	accesses := make([]*model.ProductAccess, 0)

	err := r.DB.NewSelect().
		Model(&accesses).
		Where(`product_id IN (SELECT id FROM products WHERE mini_app_id = ?)`, miniAppID).
		Where(`user_id = ?`, userID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return accesses, nil
}

func (r *ProductRepository) Feedback(ctx context.Context, productID uuid.UUID) (*model.ProductFeedback, error) {
	var feedback model.ProductFeedback

	err := r.DB.NewRaw(`
	SELECT
		ROUND(COALESCE(AVG(NULLIF(score, 0)),0))::INT AS avg_score,
		COUNT(*) AS total_reviews
	FROM reviews AS r
	JOIN lessons AS l ON l.id = r.lesson_id
	WHERE l.product_id = ?
	`, productID).
		Scan(ctx, &feedback)

	if err != nil {
		return nil, err
	}

	lessonsFeedback := make([]*model.LessonFeedback, 0)

	err = r.DB.NewRaw(`
	SELECT
		l.id AS lesson_id,
		l.module_name,
		l.content_type,
		l.title,
		ROUND(COALESCE(AVG(NULLIF(score, 0)),0))::INT AS avg_score,
		COUNT(r.id) AS total_reviews
	FROM lessons AS l
	LEFT JOIN reviews AS r ON r.lesson_id = l.id
	WHERE l.product_id = ?
	GROUP BY l.id, l.module_name, l.content_type, l.title
	ORDER BY l.index
	`, productID).
		Scan(ctx, &lessonsFeedback)

	if err != nil {
		return nil, err
	}

	feedback.Lessons = lessonsFeedback

	return &feedback, nil
}

func (r *ProductRepository) Students(
	ctx context.Context,
	productID uuid.UUID, usernameSearch string,
	limit, offset uint,
) (*model.ProductStudents, error) {

	var productStudents model.ProductStudents
	productStudents.Students = make([]*model.StudentProgress, 0)

	query1 := r.DB.NewSelect().
		With("product_lessons", r.DB.NewSelect().
			ColumnExpr(`id`).
			TableExpr(`lessons`).
			Where(`product_id = ?`, productID)).
		With("total_lessons", r.DB.NewSelect().
			ColumnExpr(`COUNT(*) AS total_lessons`).
			TableExpr(`product_lessons`)).
		ColumnExpr(`
			tl.total_lessons,
			ROUND(
				SUM(
					CASE
						WHEN lp.status = 'accepted' THEN 10000
						WHEN lp.status = 'pending' THEN 5000
						ELSE 0
					END
				)::NUMERIC / (COUNT(DISTINCT user_id) * tl.total_lessons)
			)::INT AS avg_progress`).
		TableExpr(`lesson_progress AS lp`).
		Join(`JOIN users AS u ON u.id = lp.user_id AND u.role = 'student'`).
		Join(`JOIN product_lessons AS l ON l.id = lp.lesson_id`).
		Join(`CROSS JOIN total_lessons AS tl`).
		GroupExpr(`tl.total_lessons`)

	err := query1.Scan(ctx, &productStudents)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return &productStudents, nil
	}

	query2 := r.DB.NewSelect().
		With("product_lessons", r.DB.NewSelect().
			ColumnExpr(`id`).
			TableExpr(`lessons`).
			Where(`product_id = ?`, productID)).
		With("total_lessons", r.DB.NewSelect().
			ColumnExpr(`COUNT(*) AS total_lessons`).
			TableExpr(`product_lessons`)).
		ColumnExpr(`
			lp.user_id,
			u.telegram_id,
			u.telegram_username,
			u.first_name,
			u.avatar,
			u.created_at AS joined_at,
			ROUND(
				SUM(
					CASE
						WHEN status = 'accepted' THEN 10000
						WHEN status = 'pending' THEN 5000
						ELSE 0
					END
				)::NUMERIC / tl.total_lessons
			)::INT AS progress`).
		TableExpr(`lesson_progress AS lp`).
		Join(`JOIN users AS u ON u.id = lp.user_id AND u.role = 'student'`).
		Join(`JOIN product_lessons AS l ON l.id = lp.lesson_id`).
		Join(`CROSS JOIN total_lessons AS tl`).
		GroupExpr(`tl.total_lessons, lp.user_id, u.telegram_id, u.telegram_username, u.first_name, u.avatar, u.created_at`).
		OrderExpr(`lp.user_id`).
		Limit(int(limit))

	if offset != 0 {
		query2 = query2.Offset(int(offset))
	}

	if usernameSearch != "" {
		query2 = query2.Where(`u.telegram_username ILIKE ('%' || ? || '%')`, usernameSearch)
	}

	err = query2.Scan(ctx, &productStudents.Students)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &productStudents, nil
}

func (r *ProductRepository) StudentsDetails(
	ctx context.Context,
	product *model.Product,
	dateFrom, dateTo time.Time,
) ([]*model.StudentDetails, []string, error) {

	result := make([]*model.StudentDetails, 0)

	err := r.DB.NewRaw(`
	WITH product_lessons AS (
		SELECT id FROM lessons WHERE product_id = ?
	),
	total_lessons AS (
		SELECT COUNT(*) AS total_lessons FROM product_lessons
	)
	SELECT
		u.id AS user_id,
		u.first_name,
		u.last_name,
		u.telegram_id,
		u.telegram_username,
		u.created_at AS joined_at,
		COUNT( CASE WHEN lp.status = 'accepted' THEN 1 END ) AS completed_lessons,
		tl.total_lessons
	FROM lesson_progress AS lp
	JOIN product_lessons AS l ON l.id = lp.lesson_id
	JOIN users AS u ON u.id = lp.user_id AND u.role = 'student'
	CROSS JOIN total_lessons AS tl
	WHERE lp.updated_at BETWEEN ? AND ?
	GROUP BY tl.total_lessons, u.id, u.first_name, u.last_name,
		u.telegram_id, u.telegram_username, u.created_at
	ORDER BY u.id
	`, product.ID, dateFrom, dateTo.AddDate(0, 0, 1)).
		Scan(ctx, &result)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	users := make(map[uuid.UUID]*model.StudentDetails, len(result))
	for _, s := range result {
		s.Lessons = make([]bool, len(product.Lessons))
		users[s.UserID] = s
	}

	lessonsIndex := make(map[uuid.UUID]int, len(product.Lessons))
	lessonTitles := make([]string, len(product.Lessons))
	for i, l := range product.Lessons {
		lessonsIndex[l.ID] = i
		lessonTitles[i] = l.Title
	}

	completed := []struct {
		UserID   uuid.UUID `bun:"user_id"`
		LessonID uuid.UUID `bun:"lesson_id"`
	}{}

	err = r.DB.NewRaw(`
	SELECT
		lp.user_id,
		lp.lesson_id
	FROM lessons AS l
	JOIN lesson_progress AS lp ON lp.lesson_id = l.id AND lp.updated_at BETWEEN ? AND ?
	JOIN users AS u ON u.id = lp.user_id AND u.role = 'student'
	WHERE l.product_id = ?
	`, dateFrom, dateTo, product.ID).
		Scan(ctx, &completed)

	if err != nil {
		return nil, nil, err
	}

	for _, c := range completed {
		if _, ok := users[c.UserID]; !ok {
			continue
		}
		if len(users[c.UserID].Lessons) <= lessonsIndex[c.LessonID] {
			continue
		}
		users[c.UserID].Lessons[lessonsIndex[c.LessonID]] = true
	}

	return result, lessonTitles, nil
}
