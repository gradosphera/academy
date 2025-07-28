package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type LessonRepository struct {
	repository.Generic[model.Lesson, uuid.UUID]
}

func (r *LessonRepository) WithTx(tx bun.Tx) *LessonRepository {
	return &LessonRepository{Generic: r.Generic.WithTx(tx)}
}

func NewLessonRepository(
	genericRepository repository.Generic[model.Lesson, uuid.UUID],
) *LessonRepository {
	return &LessonRepository{
		Generic: genericRepository,
	}
}

func (r *LessonRepository) Create(ctx context.Context, lesson *model.Lesson) error {
	_, err := r.DB.NewInsert().Model(lesson).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *LessonRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Lesson, error) {
	lesson := new(model.Lesson)

	query := r.DB.NewSelect().
		Model(lesson).
		Where(`id = ?`, id).
		Relation("Materials", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("category", "index")
		})

	if userID != uuid.Nil {
		query = query.
			Relation("Progress", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Where("lesson_progress.user_id = ?", userID)
			}).
			Relation("PrevLessonProgress", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Where("lesson_progress.user_id = ?", userID)
			})
	}

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	return lesson, nil
}

func (r *LessonRepository) Update(ctx context.Context, lesson *model.Lesson) error {
	_, err := r.DB.NewUpdate().
		Model(lesson).
		WherePK().
		Exec(ctx)

	return err
}

func (r *LessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var lesson *model.Lesson

	_, err := r.DB.NewDelete().
		Model(lesson).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *LessonRepository) UnlockedLessons(
	ctx context.Context,
	productID, userID uuid.UUID,
) ([]model.UnlockedLesson, error) {

	lessons := make([]model.UnlockedLesson, 0)

	err := r.DB.NewSelect().
		ColumnExpr(`paid_lessons.lesson_id, MAX(payments.access_start + payments.access_duration) AS expired_at`).
		TableExpr(`payments`).
		Join(`JOIN paid_lessons ON paid_lessons.payment_id = payments.id`).
		Where(`payments.product_id = ?`, productID).
		Where(`payments.user_id = ?`, userID).
		Where(`payments.status = ?`, model.PaymentStatusCompleted).
		GroupExpr(`paid_lessons.lesson_id`).
		Scan(ctx, &lessons)

	if err != nil {
		return nil, err
	}

	return lessons, nil
}

func (r *LessonRepository) IsLessonUnlocked(
	ctx context.Context,
	lessonID, userID uuid.UUID,
) (bool, error) {

	lessons := make([]struct {
		IsPayable bool `bun:"is_payable"`
		IsPaid    bool `bun:"is_paid"`
	}, 0)

	now := time.Now()

	err := r.DB.NewRaw(`
	SELECT
		pl.id IS NOT NULL AS is_payable,
		payments.id IS NOT NULL AS is_paid
	FROM lessons l
	LEFT JOIN product_level_lessons AS pll ON l.id = pll.lesson_id
	LEFT JOIN product_levels AS pl ON pl.id = pll.product_level_id
	LEFT JOIN payments ON payments.product_level_id = pl.id
		AND payments.user_id = ?
		AND payments.status = ?
		AND payments.access_start < CURRENT_TIMESTAMP
		AND ( payments.access_duration IS NULL OR ? < (payments.access_start + payments.access_duration) )
	LEFT JOIN paid_lessons ON paid_lessons.payment_id = payments.id AND paid_lessons.lesson_id = l.id
	WHERE l.id = ? AND (pl.id IS NULL OR paid_lessons.payment_id IS NOT NULL)
	LIMIT 1
	`, userID, model.PaymentStatusCompleted, now, lessonID).
		Scan(ctx, &lessons)

	if err != nil {
		return false, err
	}

	return len(lessons) != 0, nil
}
