package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const MaxScore = 10_000

type LessonProgressStatus string

const (
	LessonProgressStatusPending  LessonProgressStatus = "pending"
	LessonProgressStatusFailed   LessonProgressStatus = "failed"
	LessonProgressStatusAccepted LessonProgressStatus = "accepted"
)

type LessonProgress struct {
	bun.BaseModel `bun:"table:lesson_progress"`

	UserID   uuid.UUID            `bun:"user_id,pk,type:uuid,notnull" json:"user_id"`
	LessonID uuid.UUID            `bun:"lesson_id,pk,type:uuid,notnull" json:"lesson_id"`
	Status   LessonProgressStatus `bun:"status,type:lesson_progress_status,notnull" json:"status"`
	Data     json.RawMessage      `bun:"data,type:jsonb,notnull,default:'{}'" json:"data"`
	Score    int64                `bun:"score,type:int,notnull" json:"score"`
	Size     int64                `bun:"size,type:int,notnull,default:0" json:"size"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
}

func NewLessonProgress(userID, lessonID uuid.UUID) *LessonProgress {
	now := time.Now().UTC()
	return &LessonProgress{
		UserID:   userID,
		LessonID: lessonID,

		UpdatedAt: now,
		CreatedAt: now,
	}
}

func NewLessonProgressFromEmptyHomework(userID, lessonID uuid.UUID) *LessonProgress {
	now := time.Now().UTC()
	return &LessonProgress{
		UserID:   userID,
		LessonID: lessonID,
		Status:   LessonProgressStatusAccepted,
		Score:    MaxScore,

		UpdatedAt: now,
		CreatedAt: now,
	}
}

func NewLessonProgressFromQuiz(
	userID, lessonID uuid.UUID,
	data json.RawMessage,
	score int64,
) *LessonProgress {

	now := time.Now().UTC()
	return &LessonProgress{
		UserID:   userID,
		LessonID: lessonID,
		Status:   LessonProgressStatusAccepted,
		Data:     data,
		Score:    score,

		UpdatedAt: now,
		CreatedAt: now,
	}
}

func NewLessonProgressFromOpenQuestion(
	userID, lessonID uuid.UUID,
	data json.RawMessage, size int64,
) *LessonProgress {

	now := time.Now().UTC()
	return &LessonProgress{
		UserID:   userID,
		LessonID: lessonID,
		Status:   LessonProgressStatusPending,
		Data:     data,
		Size:     size,

		UpdatedAt: now,
		CreatedAt: now,
	}
}

type LessonProgressData struct {
	// Quiz only fields.
	QuizResults []QuizResult `json:"quiz_results"`

	// Open question only fields.
	OpenAnswer    string          `json:"open_answer"`
	Links         []string        `json:"links"`
	FilesMetadata []*FileMetadata `json:"files_metadata"`

	// Teacher feedback text.
	Feedback string `json:"feedback"`
}

type FileMetadata struct {
	Filename         string `json:"filename,omitempty"`
	OriginalFilename string `json:"original_filename,omitempty"`
	Size             int64  `json:"size,omitempty"`
}

type QuizResult struct {
	CorrectAnswers []bool `json:"correct_answers"`
	UserAnswers    []bool `json:"user_answers"`
}

type FilterProductHomeworkRequest struct {
	UserID   []uuid.UUID `json:"user_id"`
	LessonID []uuid.UUID `json:"lesson_id"`

	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

type FilterUserHomeworkRequest struct {
	LessonID []uuid.UUID            `json:"lesson_id"`
	Status   []LessonProgressStatus `json:"status"`

	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

type FilterLessonProgressRequest struct {
	ProductID []uuid.UUID            `json:"product_id"`
	UserID    []uuid.UUID            `json:"user_id"`
	LessonID  []uuid.UUID            `json:"lesson_id"`
	Status    []LessonProgressStatus `json:"status"`

	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

type FeedbackHomeworkRequest struct {
	LessonID  uuid.UUID            `json:"lesson_id"`
	UserID    uuid.UUID            `json:"user_id"`
	Score     int64                `json:"score"`
	Feedback  string               `json:"feedback"`
	NewStatus LessonProgressStatus `json:"new_status"`
}

type BanUserRequest struct {
	Reason    string      `json:"reason"`
	ProductID []uuid.UUID `json:"product_id"`
}

type UnbanUserRequest struct {
	ProductID []uuid.UUID `json:"product_id"`
}

type LevelUpUserRequest struct {
	ProductLevelID []uuid.UUID `json:"product_level_id"`
}

type UserLevelsRequest struct {
	ProductID uuid.UUID `json:"product_id"`
}
