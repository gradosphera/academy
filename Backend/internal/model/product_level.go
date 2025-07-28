package model

import (
	"academy/internal/types"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type ProductLevel struct {
	bun.BaseModel `bun:"table:product_levels"`

	ID          uuid.UUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ProductID   uuid.UUID       `bun:"product_id,type:uuid,notnull" json:"product_id"`
	Index       int64           `bun:"index,type:int,notnull" json:"index"`
	Name        string          `bun:"name,type:varchar(100),notnull" json:"name"`
	Description string          `bun:"description,type:text,notnull" json:"description"`
	Price       decimal.Decimal `bun:"price,type:decimal,notnull" json:"price"`
	FullPrice   decimal.Decimal `bun:"full_price,type:decimal,notnull" json:"full_price"`
	Currency    string          `bun:"currency,type:varchar(10),notnull" json:"currency"`
	Duration    types.Interval  `bun:"duration,type:interval,nullzero" json:"duration"`
	IsActive    bool            `bun:"is_active,type:boolean,notnull" json:"is_active"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	ProductLevelLessons []*ProductLevelLesson `bun:"rel:has-many,join:id=product_level_id" json:"product_level_lessons"`
	Bonus               []*Material           `bun:"rel:has-many,join:id=product_level_id" json:"bonus"`
}

func NewProductLevel() *ProductLevel {
	now := time.Now().UTC()
	return &ProductLevel{
		ID:        uuid.New(),
		UpdatedAt: now,
		CreatedAt: now,
	}
}

type ProductLevelLesson struct {
	bun.BaseModel `bun:"table:product_level_lessons"`

	ProductLevelID uuid.UUID `bun:"product_level_id,pk,type:uuid,notnull" json:"product_level_id"`
	LessonID       uuid.UUID `bun:"lesson_id,pk,type:uuid,notnull" json:"lesson_id"`
	CreatedAt      time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	Lesson *Lesson `bun:"rel:belongs-to,join:lesson_id=id" json:"lessons,omitempty"`
}

func NewProductLevelLesson(productLevelID, lessonID uuid.UUID) *ProductLevelLesson {
	now := time.Now().UTC()
	return &ProductLevelLesson{
		ProductLevelID: productLevelID,
		LessonID:       lessonID,
		CreatedAt:      now,
	}
}

type CreateProductLevelRequest struct {
	ProductID   uuid.UUID       `json:"product_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	FullPrice   decimal.Decimal `json:"full_price"`
	Currency    string          `json:"currency"`
	IsActive    bool            `json:"is_active"`

	LessonIDs []uuid.UUID `json:"lesson_ids"`
}

func (r *CreateProductLevelRequest) ToProductLevel(
	index int64, duration types.Interval,
) (*ProductLevel, error) {

	lvl := NewProductLevel()

	if r.ProductID == uuid.Nil {
		return nil, errors.New("invalid product_id")
	}

	lvl.ProductID = r.ProductID
	lvl.Index = index
	lvl.Name = r.Name
	lvl.Description = r.Description
	lvl.Price = r.Price
	lvl.FullPrice = r.FullPrice
	lvl.Currency = r.Currency
	lvl.Duration = duration
	lvl.IsActive = r.IsActive

	return lvl, nil
}

type BuyProductLevelRequest struct {
	ProductLevelID uuid.UUID `json:"product_level_id"`
}

type EditProductLevelRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       decimal.Decimal `json:"price"`
	FullPrice   decimal.Decimal `json:"full_price"`
	Currency    string          `json:"currency"`
	IsActive    bool            `json:"is_active"`

	AddLessonIDs    []uuid.UUID `json:"add_lesson_ids"`
	RemoveLessonIDs []uuid.UUID `json:"remove_lesson_ids"`
}

func (r *EditProductLevelRequest) UpdateProductLevel(productLevel *ProductLevel) (bool, error) {
	isChanged := false

	if r.Name != productLevel.Name {
		productLevel.Name = r.Name
		isChanged = true
	}
	if r.Description != productLevel.Description {
		productLevel.Description = r.Description
		isChanged = true
	}
	if !r.Price.Equal(productLevel.Price) {
		productLevel.Price = r.Price
		isChanged = true
	}
	if !r.FullPrice.Equal(productLevel.FullPrice) {
		productLevel.FullPrice = r.FullPrice
		isChanged = true
	}
	if r.Currency != productLevel.Currency {
		productLevel.Currency = r.Currency
		isChanged = true
	}
	if r.IsActive != productLevel.IsActive {
		productLevel.IsActive = r.IsActive
		isChanged = true
	}

	productLevel.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}
