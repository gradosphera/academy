package model

import (
	"academy/internal/types"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type PaymentStatus string

const (
	PaymentStatusPending       PaymentStatus = "pending"
	PaymentStatusCompleted     PaymentStatus = "completed"
	PaymentStatusFailed        PaymentStatus = "failed"
	PaymentStatusRefunded      PaymentStatus = "refunded"
	PaymentStatusPendingRefund PaymentStatus = "pending_refund"
)

type Payment struct {
	bun.BaseModel `bun:"table:payments"`

	ID             uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	MiniAppID      uuid.UUID `bun:"mini_app_id,type:uuid,notnull" json:"-"`
	ProductID      uuid.UUID `bun:"product_id,type:uuid,nullzero" json:"-"`
	UserID         uuid.UUID `bun:"user_id,type:uuid,nullzero" json:"user_id"`
	PlanID         uuid.UUID `bun:"plan_id,type:uuid,nullzero" json:"plan_id,omitempty"`
	ProductLevelID uuid.UUID `bun:"product_level_id,type:uuid,nullzero" json:"product_level_id,omitempty"`

	AccessStart    types.Time      `bun:"access_start,type:timestamptz,notnull" json:"access_start"`
	AccessDuration types.Interval  `bun:"access_duration,type:interval,nullzero" json:"access_duration"`
	Amount         decimal.Decimal `bun:"amount,type:decimal,notnull" json:"amount"`
	Currency       string          `bun:"currency,type:varchar(10),notnull" json:"currency"`
	AmountUSD      decimal.Decimal `bun:"amount_usd,type:decimal(12,2),notnull" json:"amount_usd"`
	Status         PaymentStatus   `bun:"status,type:payment_status,notnull" json:"status"`
	URL            string          `bun:"url,type:varchar(255),notnull" json:"url"`
	Comment        string          `bun:"comment,type:text,notnull,default:''" json:"comment"`
	UpdatedAt      time.Time       `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt      time.Time       `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	MiniApp      *MiniApp      `bun:"rel:belongs-to,join:mini_app_id=id" json:"-"`
	User         *User         `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Plan         *Plan         `bun:"rel:belongs-to,join:plan_id=id" json:"plan,omitempty"`
	ProductLevel *ProductLevel `bun:"rel:belongs-to,join:product_level_id=id" json:"product_level,omitempty"`
}

func NewPaymentForProductLevel(userID uuid.UUID, product *Product, productLevel *ProductLevel) *Payment {
	now := time.Now().UTC()

	accessStart := types.NewTime(now)
	if product.ReleaseDate.Valid && accessStart.Time.Before(product.ReleaseDate.Time) {
		accessStart = product.ReleaseDate
	}

	return &Payment{
		ID:             uuid.New(),
		MiniAppID:      product.MiniAppID,
		ProductID:      product.ID,
		UserID:         userID,
		ProductLevelID: productLevel.ID,

		AccessStart:    accessStart,
		AccessDuration: productLevel.Duration,
		Amount:         productLevel.Price.RoundDown(2),
		Currency:       productLevel.Currency,
		Status:         PaymentStatusPending,
		Comment:        fmt.Sprintf("%s - %s", product.Title, productLevel.Name),

		UpdatedAt: now,
		CreatedAt: now,
	}
}

func NewFreePaymentForProductLevel(userID uuid.UUID, product *Product, productLevel *ProductLevel) *Payment {
	now := time.Now().UTC()

	accessStart := types.NewTime(now)
	if product.ReleaseDate.Valid && accessStart.Time.Before(product.ReleaseDate.Time) {
		accessStart = product.ReleaseDate
	}

	return &Payment{
		ID:             uuid.New(),
		MiniAppID:      product.MiniAppID,
		ProductID:      product.ID,
		UserID:         userID,
		ProductLevelID: productLevel.ID,

		AccessStart:    accessStart,
		AccessDuration: productLevel.Duration,
		Status:         PaymentStatusCompleted,
		Comment:        fmt.Sprintf("%s - %s", product.Title, productLevel.Name),

		UpdatedAt: now,
		CreatedAt: now,
	}
}

type GetPaymentsRequest struct {
	Status []PaymentStatus `json:"status"`

	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

type FilterPayments struct {
	MiniAppID uuid.UUID
	ID        []uuid.UUID
	UserID    []uuid.UUID
	Status    []PaymentStatus

	Limit  uint
	Offset uint
}

type ExportStudentsPaymentsRequest struct {
	DateFrom string `json:"date_from"`
	DateTo   string `json:"date_to"`
}

type GetStudentsPaymentsRequest struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

type StudentsPayment struct {
	PaymentID        uuid.UUID `bun:"payment_id"`
	UserID           uuid.UUID `bun:"user_id"`
	TelegramID       int64     `bun:"telegram_id"`
	TelegramUsername string    `bun:"telegram_username"`

	ProductName      string `bun:"product_name"`
	ProductLevelName string `bun:"product_level_name"`

	AmountUSD string `bun:"amount_usd"`

	PaidAt time.Time `bun:"paid_at"`
}
