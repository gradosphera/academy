package model

import (
	"academy/internal/types"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type LessonType string

const (
	LessonTypeVideo LessonType = "video"
	LessonTypeAudio LessonType = "audio"
	LessonTypeText  LessonType = "text"
	LessonTypeEvent LessonType = "event"
)

type Lesson struct {
	bun.BaseModel `bun:"table:lessons"`

	ID               uuid.UUID      `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ProductID        uuid.UUID      `bun:"product_id,type:uuid,notnull" json:"product_id"`
	Index            int64          `bun:"index,type:int,notnull" json:"index"`
	ModuleName       string         `bun:"module_name,type:varchar(100),notnull" json:"module_name"`
	ContentType      LessonType     `bun:"content_type,type:lesson_type,notnull" json:"content_type"`
	Title            string         `bun:"title,type:varchar(100),notnull" json:"title"`
	Description      string         `bun:"description,type:text,notnull" json:"description"`
	PreviousLessonID uuid.UUID      `bun:"previous_lesson_id,type:uuid,nullzero" json:"previous_lesson_id"`
	ReleaseDate      types.Time     `bun:"release_date,type:timestamptz,nullzero" json:"release_date"`
	AccessTime       types.Interval `bun:"access_time,type:interval,nullzero" json:"access_time"`
	IsActive         bool           `bun:"is_active,type:boolean,notnull" json:"is_active"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	PreviousLesson *Lesson `bun:"rel:belongs-to,join:previous_lesson_id=id" json:"previous_lesson,omitempty"`

	Materials []*Material `bun:"rel:has-many,join:id=lesson_id" json:"materials"`

	Progress           []*LessonProgress `bun:"rel:has-many,join:id=lesson_id" json:"progress,omitempty"`
	PrevLessonProgress []*LessonProgress `bun:"rel:has-many,join:previous_lesson_id=lesson_id" json:"previous_lesson_progress,omitempty"`
}

func NewLesson() *Lesson {
	now := time.Now().UTC()
	return &Lesson{
		ID:        uuid.New(),
		UpdatedAt: now,
		CreatedAt: now,
	}
}

type UnlockedLesson struct {
	LessonID  uuid.UUID `bun:"lesson_id"`
	ExpiredAT time.Time `bun:"expired_at"`
}

type CreateLessonRequest struct {
	ProductID   uuid.UUID      `json:"product_id"`
	ModuleName  string         `json:"module_name"`
	ContentType LessonType     `json:"content_type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ReleaseDate types.Time     `json:"release_date"`
	AccessTime  types.Interval `json:"access_time"`
	IsActive    bool           `json:"is_active"`

	Index *int64 `json:"index"`

	ProductLevelID []uuid.UUID `json:"product_level_id"`
}

func (r *CreateLessonRequest) ToLesson() (*Lesson, error) {
	l := NewLesson()

	if r.ProductID == uuid.Nil {
		return nil, errors.New("invalid product_id")
	}

	l.ProductID = r.ProductID
	l.ModuleName = r.ModuleName
	l.Title = r.Title
	l.Description = r.Description
	l.ContentType = r.ContentType

	l.ReleaseDate = r.ReleaseDate
	l.AccessTime = r.AccessTime
	l.IsActive = r.IsActive

	return l, nil
}

type EditLessonRequest struct {
	ModuleName  string         `json:"module_name"`
	ContentType LessonType     `json:"content_type"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ReleaseDate types.Time     `json:"release_date"`
	AccessTime  types.Interval `json:"access_time"`
	IsActive    bool           `json:"is_active"`
}

func (r *EditLessonRequest) UpdateLesson(l *Lesson) (bool, error) {
	isChanged := false

	if r.ModuleName != l.ModuleName {
		l.ModuleName = r.ModuleName
		isChanged = true
	}
	if r.ContentType != l.ContentType {
		l.ContentType = r.ContentType
		isChanged = true
	}
	if r.Title != l.Title {
		l.Title = r.Title
		isChanged = true
	}
	if r.Description != l.Description {
		l.Description = r.Description
		isChanged = true
	}
	if !r.ReleaseDate.IsEqual(l.ReleaseDate) {
		l.ReleaseDate = r.ReleaseDate
		isChanged = true
	}
	if !r.AccessTime.IsEqual(l.AccessTime) {
		l.AccessTime = r.AccessTime
		isChanged = true
	}
	if r.IsActive != l.IsActive {
		l.IsActive = r.IsActive
		isChanged = true
	}

	l.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}

type LessonSubmitionRequest struct {
	QuizAnswers [][]bool `json:"quiz"`

	Text  string   `json:"text"`
	Links []string `json:"links"`
}

type LessonSubmitionResponce struct {
	LessonResult *LessonProgress `json:"lesson_result"`
}

type QuestionSubmitionRequest struct {
	QuestionIndex  int    `json:"question_index"`
	QuestionAnswer []bool `json:"question_answer"`
}

type QuestionSubmitionResponce struct {
	QuestionResult *QuizResult `json:"question_result"`
}
