package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Review struct {
	bun.BaseModel `bun:"table:reviews"`

	ID       uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	UserID   uuid.UUID `bun:"user_id,type:uuid,notnull" json:"user_id"`
	LessonID uuid.UUID `bun:"lesson_id,type:uuid,notnull" json:"lesson_id"`
	Score    int64     `bun:"score,type:int,notnull" json:"score"`

	Text string `bun:"text,type:text,notnull" json:"text"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
}

func NewReview() *Review {
	now := time.Now().UTC()
	return &Review{
		ID:        uuid.New(),
		UpdatedAt: now,
		CreatedAt: now,
	}
}

type ReviewRequest struct {
	Score int64  `json:"score"`
	Text  string `json:"text"`
}

func (r *ReviewRequest) Validate() error {
	if r.Score < 0 || MaxScore < r.Score {
		return fmt.Errorf("invalid score: %v", r.Score)
	}

	return nil
}

func NewLessonReview(userID, lessonID uuid.UUID, req *ReviewRequest) *Review {
	now := time.Now().UTC()
	return &Review{
		ID:        uuid.New(),
		UserID:    userID,
		LessonID:  lessonID,
		Score:     req.Score,
		Text:      req.Text,
		UpdatedAt: now,
		CreatedAt: now,
	}
}
