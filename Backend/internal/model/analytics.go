package model

import (
	"academy/internal/types"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

var allowedPeriods = map[pgtype.Interval]struct{}{
	{Days: 7, Valid: true}:    {},
	{Months: 1, Valid: true}:  {},
	{Months: 12, Valid: true}: {},
}

type AnalyticsRequest struct {
	TimePeriod types.Interval `json:"time_period"`
}

func (r *AnalyticsRequest) Validate() error {
	_, ok := allowedPeriods[r.TimePeriod.Interval]
	if !ok {
		return fmt.Errorf("time_period is not allowed")
	}

	return nil
}

// Mini App Analytics.

type Analytics struct {
	NewStudents   int64 `bun:"new_students" json:"new_students"`
	TotalStudents int64 `bun:"total_students" json:"total_students"`

	MoneyEarned      decimal.Decimal `bun:"money_earned" json:"money_earned"`
	TotalMoneyEarned decimal.Decimal `bun:"total_money_earned" json:"total_money_earned"`

	Products []*ProductAnalytics `bun:"-" json:"products"`
}

type ProductAnalytics struct {
	ProductID uuid.UUID `bun:"product_id" json:"product_id"`

	NewStudents   int64 `bun:"new_students" json:"new_students"`
	TotalStudents int64 `bun:"total_students" json:"total_students"`

	MoneyEarned      decimal.Decimal `bun:"money_earned" json:"money_earned"`
	TotalMoneyEarned decimal.Decimal `bun:"total_money_earned" json:"total_money_earned"`
}

// Product Analytics. Page 1. Feedback.

type ProductFeedback struct {
	AvgScore     int64 `bun:"avg_score" json:"avg_score"`
	TotalReviews int64 `bun:"total_reviews" json:"total_reviews"`

	Lessons []*LessonFeedback `bun:"-" json:"lessons"`
}

type LessonFeedback struct {
	LessonID    uuid.UUID  `bun:"lesson_id" json:"lesson_id"`
	ModuleName  string     `bun:"module_name" json:"module_name"`
	ContentType LessonType `bun:"content_type" json:"content_type"`
	Title       string     `bun:"title" json:"title"`

	AvgScore     int64 `bun:"avg_score" json:"avg_score"`
	TotalReviews int64 `bun:"total_reviews" json:"total_reviews"`
}

// Product Analytics. Page 2. Students.

type ProductStudents struct {
	AvgProgress  int64 `bun:"avg_progress" json:"avg_progress"`
	TotalLessons int64 `bun:"total_lessons" json:"total_lessons"`

	Students []*StudentProgress `bun:"-" json:"students_progress"`
}

type StudentProgress struct {
	UserID           uuid.UUID `bun:"user_id" json:"user_id"`
	TelegramID       int64     `bun:"telegram_id" json:"telegram_id"`
	TelegramUsername string    `bun:"telegram_username" json:"telegram_username"`
	FirstName        string    `bun:"first_name" json:"first_name"`
	Avatar           string    `bun:"avatar" json:"avatar"`
	JoinedAt         time.Time `bun:"joined_at" json:"joined_at"`

	Progress int64 `bun:"progress" json:"progress"`
}

// Product Analytics. Page 2. Students. Export to excel.

type ExportProductStudentsRequest struct {
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

type StudentDetails struct {
	UserID           uuid.UUID `bun:"user_id"`
	FirstName        string    `bun:"first_name"`
	LastName         string    `bun:"last_name"`
	TelegramID       int64     `bun:"telegram_id"`
	TelegramUsername string    `bun:"telegram_username"`
	JoinedAt         time.Time `bun:"joined_at"`

	CompletedLessons int64 `bun:"completed_lessons"`
	TotalLessons     int64 `bun:"total_lessons"`

	Lessons []bool `bun:"-"`
}

// Product Analytics. Student Progress.

type StudentStats struct {
	ProductID          uuid.UUID `bun:"product_id"`
	ProductTitle       string    `bun:"product_title"`
	ProductCover       string    `bun:"product_cover"`
	ProductContentType string    `bun:"product_content_type"`
	JoinedAt           time.Time `bun:"joined_at"`

	ProductAccessDeletedReason string     `bun:"product_access_deleted_reason"`
	ProductAccessDeletedAt     *time.Time `bun:"product_access_deleted_at"`

	LessonID          uuid.UUID  `bun:"lesson_id"`
	ModuleName        string     `bun:"module_name"`
	LessonContentType LessonType `bun:"lesson_content_type"`
	LessonTitle       string     `bun:"lesson_title"`

	ProgressStatus LessonProgressStatus `bun:"progress_status"`
	Score          int64                `bun:"score"`

	ReviewScore int64  `bun:"review_score"`
	ReviewText  string `bun:"review_text"`
}

type StudentProductStats struct {
	ProductID   uuid.UUID `json:"product_id"`
	Title       string    `json:"title"`
	Cover       string    `json:"cover"`
	ContentType string    `json:"content_type"`
	JoinedAt    time.Time `json:"joined_at"`

	ProductAccessDeletedReason string     `json:"product_access_deleted_reason"`
	ProductAccessDeletedAt     *time.Time `json:"product_access_deleted_at"`

	LessonStats []*StudentLessonStats `json:"lesson_stats"`
}

type StudentProductLevels struct {
	PaymentID               uuid.UUID       `bun:"payment_id" json:"payment_id"`
	PaymentUpdatedAt        time.Time       `bun:"payment_updated_at" json:"payment_updated_at"`
	ProductLevelID          uuid.UUID       `bun:"product_level_id" json:"product_level_id"`
	ProductLevelName        string          `bun:"product_level_name" json:"product_level_name"`
	ProductLevelDescription string          `bun:"product_level_description" json:"product_level_description"`
	PaidPrice               decimal.Decimal `bun:"paid_price" json:"paid_price"`
	PaidCurrency            string          `bun:"paid_currency" json:"paid_currency"`
	EndsAt                  time.Time       `bun:"ends_at" json:"ends_at"`
}

type StudentLessonStats struct {
	LessonID    uuid.UUID  `json:"lesson_id"`
	ModuleName  string     `json:"module_name"`
	ContentType LessonType `json:"content_type"`
	Title       string     `json:"title"`

	ProgressStatus LessonProgressStatus `json:"progress_status,omitempty"`
	Score          int64                `json:"score,omitempty"`

	ReviewScore int64  `json:"review_score,omitempty"`
	ReviewText  string `json:"review_text,omitempty"`
}
