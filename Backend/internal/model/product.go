package model

import (
	"academy/internal/types"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type LessonAccess string

const (
	LessonAccessUnlocked   LessonAccess = "unlocked"
	LessonAccessSequential LessonAccess = "sequential"
	LessonAccessScheduled  LessonAccess = "scheduled"
)

type Product struct {
	bun.BaseModel `bun:"table:products"`

	ID           uuid.UUID      `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	MiniAppID    uuid.UUID      `bun:"mini_app_id,type:uuid,notnull" json:"mini_app_id"`
	Index        int64          `bun:"index,type:int,notnull" json:"index"`
	Title        string         `bun:"title,type:varchar(100),notnull" json:"title"`
	Cover        string         `bun:"cover,type:varchar(255),notnull" json:"cover"`
	CoverSize    int64          `bun:"cover_size,type:bigint,notnull" json:"cover_size"`
	Description  string         `bun:"description,type:text,notnull" json:"description"`
	ContentType  string         `bun:"content_type,type:varchar(100),notnull" json:"content_type"`
	Tags         []string       `bun:"tags,array,type:varchar(100)[],notnull" json:"tags"`
	LessonAccess LessonAccess   `bun:"lesson_access,type:lesson_access,notnull" json:"lesson_access"`
	ReleaseDate  types.Time     `bun:"release_date,type:timestamptz,nullzero" json:"release_date"`
	AccessTime   types.Interval `bun:"access_time,type:interval,nullzero" json:"access_time"`
	IsActive     bool           `bun:"is_active,type:boolean,notnull" json:"is_active"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	Lessons []*Lesson       `bun:"rel:has-many,join:id=product_id" json:"lessons,omitempty"`
	Levels  []*ProductLevel `bun:"rel:has-many,join:id=product_id" json:"product_levels,omitempty"`
}

func NewProduct() *Product {
	now := time.Now().UTC()
	return &Product{
		ID:        uuid.New(),
		UpdatedAt: now,
		CreatedAt: now,
	}
}

type CreateProductRequest struct {
	Index        int64          `json:"index"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	ContentType  string         `json:"content_type"`
	Tags         []string       `json:"tags"`
	LessonAccess LessonAccess   `json:"lesson_access"`
	ReleaseDate  types.Time     `json:"release_date"`
	AccessTime   types.Interval `json:"access_time"`
	IsActive     bool           `json:"is_active"`
}

func (r *CreateProductRequest) ToProduct(miniAppID uuid.UUID) (*Product, error) {
	p := NewProduct()

	p.MiniAppID = miniAppID
	p.Index = r.Index
	p.Title = r.Title
	p.Description = r.Description
	p.ContentType = r.ContentType
	p.Tags = r.Tags
	if p.Tags == nil {
		p.Tags = []string{}
	}
	p.LessonAccess = r.LessonAccess
	p.ReleaseDate = r.ReleaseDate
	p.AccessTime = r.AccessTime
	p.IsActive = r.IsActive

	return p, nil
}

type EditProductRequest struct {
	Index        int64          `json:"index"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	ContentType  string         `json:"content_type"`
	Tags         []string       `json:"tags"`
	LessonAccess LessonAccess   `json:"lesson_access"`
	ReleaseDate  types.Time     `json:"release_date"`
	AccessTime   types.Interval `json:"access_time"`
	IsActive     bool           `json:"is_active"`
	DeleteCover  bool           `json:"delete_cover"`
}

func (r *EditProductRequest) UpdateProduct(p *Product) (bool, error) {
	isChanged := false

	if r.Index != p.Index {
		p.Index = r.Index
		isChanged = true
	}
	if r.Title != p.Title {
		p.Title = r.Title
		isChanged = true
	}
	if r.Description != p.Description {
		p.Description = r.Description
		isChanged = true
	}
	if r.ContentType != p.ContentType {
		p.ContentType = r.ContentType
		isChanged = true
	}
	if slices.Compare(r.Tags, p.Tags) != 0 {
		p.Tags = r.Tags
		isChanged = true
	}
	if r.LessonAccess != p.LessonAccess {
		p.LessonAccess = r.LessonAccess
		isChanged = true
	}
	if !r.ReleaseDate.IsEqual(p.ReleaseDate) {
		p.ReleaseDate = r.ReleaseDate
		isChanged = true
	}
	if !r.AccessTime.IsEqual(p.AccessTime) {
		p.AccessTime = r.AccessTime
		isChanged = true
	}
	if r.IsActive != p.IsActive {
		p.IsActive = r.IsActive
		isChanged = true
	}

	p.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}

type ReorderProductLessonsRequest struct {
	Modules []struct {
		ModuleName string      `json:"module_name"`
		LessonIDs  []uuid.UUID `json:"lesson_ids"`
	} `json:"modules"`

	LessonsToDelete []uuid.UUID `json:"lessons_to_delete"`
}

func (r *ReorderProductLessonsRequest) Validate() error {
	lessons := make(map[uuid.UUID]struct{}, len(r.Modules))

	for _, m := range r.Modules {
		if len(m.LessonIDs) == 0 {
			return fmt.Errorf("module should include at least one lesson")
		}
		for _, lessonID := range m.LessonIDs {
			if lessonID == uuid.Nil {
				return fmt.Errorf("invalid lesson id")
			}

			if _, ok := lessons[lessonID]; ok {
				return fmt.Errorf("duplicated lesson id")
			}

			lessons[lessonID] = struct{}{}
		}
	}

	for _, lessonID := range r.LessonsToDelete {
		if lessonID == uuid.Nil {
			return fmt.Errorf("invalid lesson id")
		}

		if _, ok := lessons[lessonID]; ok {
			return fmt.Errorf("duplicated lesson id")
		}

		lessons[lessonID] = struct{}{}
	}

	return nil
}

type ReorderProductLevelsRequest struct {
	ProductLevels []uuid.UUID `json:"product_levels"`
}

func (r *ReorderProductLevelsRequest) Validate() error {
	levels := make(map[uuid.UUID]struct{}, len(r.ProductLevels))
	for _, id := range r.ProductLevels {
		if id == uuid.Nil {
			return fmt.Errorf("invalid product level id")
		}

		if _, ok := levels[id]; ok {
			return fmt.Errorf("duplicated product level id")
		}

		levels[id] = struct{}{}
	}

	return nil
}
