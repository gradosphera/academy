package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type MaterialCategory string

const (
	MaterialCategoryLessonContent MaterialCategory = "lesson_content"
	MaterialCategoryLessonCover   MaterialCategory = "lesson_cover"
	MaterialCategoryMaterials     MaterialCategory = "materials"
	MaterialCategoryHomework      MaterialCategory = "homework"
	MaterialCategorySlides        MaterialCategory = "slides"
	MaterialCategoryTOS           MaterialCategory = "tos"
	MaterialCategoryPrivacyPolicy MaterialCategory = "privacy_policy"
	MaterialCategoryBonus         MaterialCategory = "bonus"
)

type MaterialType string

const (
	MaterialTypeCircleVideo  MaterialType = "circle_video"
	MaterialTypeVideo        MaterialType = "video"
	MaterialTypeAudio        MaterialType = "audio"
	MaterialTypePicture      MaterialType = "picture"
	MaterialTypeText         MaterialType = "text"
	MaterialTypeQuiz         MaterialType = "quiz"
	MaterialTypeOpenQuestion MaterialType = "open_question"
)

type MaterialStatus string

const (
	MaterialStatusReady              MaterialStatus = "ready"
	MaterialStatusPendingCompressing MaterialStatus = "pending_compressing"
	MaterialStatusPendingMoveToMux   MaterialStatus = "pending_move_to_mux"
)

type Material struct {
	bun.BaseModel `bun:"table:materials"`

	ID               uuid.UUID        `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	MiniAppID        uuid.UUID        `bun:"mini_app_id,type:uuid,nullzero" json:"mini_app_id,omitempty"`
	LessonID         uuid.UUID        `bun:"lesson_id,type:uuid,nullzero" json:"lesson_id,omitempty"`
	ProductLevelID   uuid.UUID        `bun:"product_level_id,type:uuid,nullzero" json:"product_level_id,omitempty"`
	Index            int64            `bun:"index,type:int,notnull" json:"index"`
	Category         MaterialCategory `bun:"category,type:material_category,notnull" json:"category"`
	ContentType      MaterialType     `bun:"content_type,type:material_type,notnull" json:"content_type"`
	Title            string           `bun:"title,type:varchar(100),notnull" json:"title"`
	Description      string           `bun:"description,type:text,notnull" json:"description"`
	URL              string           `bun:"url,type:varchar(255),notnull" json:"url"`
	Size             int64            `bun:"size,type:int,notnull,default:0" json:"size"`
	OriginalFilename string           `bun:"original_filename,type:varchar(100),notnull" json:"original_filename"`
	Filename         string           `bun:"filename,type:varchar(255),notnull" json:"filename"`
	Metadata         json.RawMessage  `bun:"metadata,type:jsonb,nullzero" json:"metadata"`
	HiddenMetadata   json.RawMessage  `bun:"hidden_metadata,type:jsonb,nullzero" json:"-"`
	Status           MaterialStatus   `bun:"status,type:material_status,nullzero,notnull,default:'ready'" json:"status"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`
}

func NewMaterial() *Material {
	now := time.Now().UTC()
	return &Material{
		ID:        uuid.New(),
		UpdatedAt: now,
		CreatedAt: now,
	}
}

type CreateMaterialRequest struct {
	MiniAppID      uuid.UUID `json:"mini_app_id"`
	LessonID       uuid.UUID `json:"lesson_id"`
	ProductLevelID uuid.UUID `json:"product_level_id"`

	Index       int64            `json:"index"`
	Category    MaterialCategory `json:"category"`
	ContentType MaterialType     `json:"content_type"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	URL         string           `json:"url"`
}

func (r *CreateMaterialRequest) ToMaterial(originalFilename, filename string, size int64) (*Material, error) {
	p := NewMaterial()

	p.MiniAppID = r.MiniAppID
	p.LessonID = r.LessonID
	p.ProductLevelID = r.ProductLevelID

	p.Index = r.Index
	p.Category = r.Category
	p.ContentType = r.ContentType
	p.Title = r.Title
	p.Description = r.Description
	p.URL = r.URL

	p.OriginalFilename = originalFilename
	p.Filename = filename
	p.Size = size

	return p, nil
}

type EditMaterialRequest struct {
	Index            int64        `json:"index"`
	ContentType      MaterialType `json:"content_type"`
	Title            string       `json:"title"`
	Description      string       `json:"description"`
	OriginalFilename string       `json:"original_filename"`
	URL              string       `json:"url"`
}

func (r *EditMaterialRequest) UpdateMaterial(material *Material) (bool, error) {
	isChanged := false

	if r.Index != material.Index {
		material.Index = r.Index
		isChanged = true
	}
	if r.ContentType != material.ContentType {
		switch material.ContentType {
		case MaterialTypeOpenQuestion, MaterialTypeQuiz:
			return false, fmt.Errorf("can't change content_type from homework type")
		}
		switch r.ContentType {
		case MaterialTypeOpenQuestion, MaterialTypeQuiz:
			return false, fmt.Errorf("can't change content_type to homework type")
		}
		material.ContentType = r.ContentType
		isChanged = true
	}
	if r.Title != material.Title {
		material.Title = r.Title
		isChanged = true
	}
	if r.Description != material.Description {
		material.Description = r.Description
		isChanged = true
	}
	if r.OriginalFilename != material.OriginalFilename {
		material.OriginalFilename = r.OriginalFilename
		isChanged = true
	}
	if r.URL != material.URL {
		material.URL = r.URL
		isChanged = true
	}

	if isChanged {
		material.UpdatedAt = time.Now().UTC()
	}

	return isChanged, nil
}

type SubmitChunksRequest struct {
	OriginalFilename string         `json:"original_filename"`
	Status           MaterialStatus `json:"status"`
}

func (r *SubmitChunksRequest) Validate() error {
	if r.OriginalFilename == "" {
		return fmt.Errorf("filename not provided")
	}

	switch r.Status {
	case "":
	case MaterialStatusReady:
	case MaterialStatusPendingCompressing:
	case MaterialStatusPendingMoveToMux:
	default:
		return fmt.Errorf("invalid status")
	}

	return nil
}

type CreateHomeworkRequest struct {
	LessonID    uuid.UUID `json:"lesson_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`

	HomeworkType MaterialType `json:"homework_type"`

	QuizMetadata *QuizMetadata       `json:"quiz_metadata"`
	QuizAnswers  *QuizHiddenMetadata `json:"quiz_answers"`

	OpenQuestionMetadata *OpenQuestionMetadata `json:"open_question_metadata"`
}

func (r *CreateHomeworkRequest) ToMaterial(questionLimit, optionLimit int) (*Material, error) {
	metadata, hiddenMetadata, err := createHomeworkMetadata(
		r.HomeworkType,
		r.QuizMetadata, r.QuizAnswers,
		r.OpenQuestionMetadata,
		questionLimit, optionLimit,
	)
	if err != nil {
		return nil, err
	}

	material := NewMaterial()
	material.LessonID = r.LessonID
	material.Category = MaterialCategoryHomework
	material.ContentType = r.HomeworkType
	material.Title = r.Title
	material.Description = r.Description
	material.Metadata = metadata
	material.HiddenMetadata = hiddenMetadata

	return material, nil
}

type EditHomeworkRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`

	HomeworkType MaterialType `json:"homework_type"`

	QuizMetadata *QuizMetadata       `json:"quiz_metadata"`
	QuizAnswers  *QuizHiddenMetadata `json:"quiz_answers"`

	OpenQuestionMetadata *OpenQuestionMetadata `json:"open_question_metadata"`
}

func (r *EditHomeworkRequest) UpdateMaterial(material *Material, questionLimit, optionLimit int) (bool, error) {
	isChanged := false

	if r.Title != material.Title {
		material.Title = r.Title
		isChanged = true
	}
	if r.Description != material.Description {
		material.Description = r.Description
		isChanged = true
	}

	metadata, hiddenMetadata, err := createHomeworkMetadata(
		r.HomeworkType,
		r.QuizMetadata, r.QuizAnswers,
		r.OpenQuestionMetadata,
		questionLimit, optionLimit,
	)
	if err != nil {
		return false, err
	}

	{
		var oldMetadata, newMetadata any
		if err := json.Unmarshal(metadata, &newMetadata); err != nil {
			return false, err
		}
		if err := json.Unmarshal(material.Metadata, &oldMetadata); err != nil {
			return false, err
		}
		if !reflect.DeepEqual(oldMetadata, newMetadata) {
			material.Metadata = metadata
			isChanged = true
		}
	}

	{
		var oldHiddenMetadata, newHiddenMetadata any
		if err := json.Unmarshal(hiddenMetadata, &newHiddenMetadata); err != nil {
			return false, err
		}
		if err := json.Unmarshal(material.HiddenMetadata, &oldHiddenMetadata); err != nil {
			return false, err
		}
		if !reflect.DeepEqual(oldHiddenMetadata, newHiddenMetadata) {
			material.HiddenMetadata = hiddenMetadata
			isChanged = true
		}
	}

	if isChanged {
		material.UpdatedAt = time.Now().UTC()
	}

	return isChanged, nil
}

func createHomeworkMetadata(homeworkType MaterialType,
	quizMetadata *QuizMetadata,
	quizAnswers *QuizHiddenMetadata,
	openQuestionMetadata *OpenQuestionMetadata,
	questionLimit, optionLimit int,
) (json.RawMessage, json.RawMessage, error) {

	var metadata json.RawMessage
	var hiddenMetadata json.RawMessage

	switch homeworkType {
	case MaterialTypeQuiz:
		if quizMetadata == nil || quizAnswers == nil || openQuestionMetadata != nil {
			return nil, nil, fmt.Errorf("unexpected request")
		}

		if len(quizMetadata.Questions) != len(quizAnswers.Answers) {
			return nil, nil, fmt.Errorf("number of answers not match questions")
		}

		for i, question := range quizMetadata.Questions {
			if question.Question == "" {
				return nil, nil, fmt.Errorf("no question provided")
			}
			if questionLimit < utf8.RuneCountInString(question.Question) {
				return nil, nil, fmt.Errorf("question exceeds the limit")
			}

			if len(question.Options) != len(quizAnswers.Answers[i]) {
				return nil, nil, fmt.Errorf("number of question options not match answers")
			}

			for _, option := range question.Options {
				if optionLimit < utf8.RuneCountInString(option) {
					return nil, nil, fmt.Errorf("question's option exceeds the limit")
				}
			}

			numberOfTrueAnswers := 0
			for _, answer := range quizAnswers.Answers[i] {
				if answer {
					numberOfTrueAnswers++
				}
			}

			switch question.AnswerType {
			case QuizAnswerTypeEmpty:
				if len(question.Options) != 0 {
					return nil, nil, fmt.Errorf("unexpected number of options")
				}
			case QuizAnswerTypeSingle:
				if len(question.Options) < 2 {
					return nil, nil, fmt.Errorf("unexpected number of options")
				}

				if numberOfTrueAnswers != 1 {
					return nil, nil, fmt.Errorf("unexpected number of answers")
				}
			case QuizAnswerTypeMulti:
				if len(question.Options) < 2 {
					return nil, nil, fmt.Errorf("unexpected number of options")
				}

				if numberOfTrueAnswers == 0 {
					return nil, nil, fmt.Errorf("no true answers")
				}
			default:
				return nil, nil, fmt.Errorf("unexpected quiz answer type: %v", question.AnswerType)
			}
		}

		rawMetadata, err := json.Marshal(quizMetadata)
		if err != nil {
			return nil, nil, fmt.Errorf("json.Marshal: %w", err)
		}
		metadata = rawMetadata

		rawHiddenMetadata, err := json.Marshal(quizAnswers)
		if err != nil {
			return nil, nil, fmt.Errorf("json.Marshal: %w", err)
		}
		hiddenMetadata = rawHiddenMetadata

	case MaterialTypeOpenQuestion:
		if openQuestionMetadata == nil || quizMetadata != nil || quizAnswers != nil {
			return nil, nil, fmt.Errorf("unexpected request")
		}

		if openQuestionMetadata.Question == "" {
			return nil, nil, fmt.Errorf("no question provided")
		}

		if questionLimit < utf8.RuneCountInString(openQuestionMetadata.Question) {
			return nil, nil, fmt.Errorf("question exceeds the limit")
		}

		b, err := json.Marshal(openQuestionMetadata)
		if err != nil {
			return nil, nil, fmt.Errorf("json.Marshal: %w", err)
		}
		metadata = b

		hiddenMetadata = nil

	default:
		return nil, nil, fmt.Errorf("unsupported homework type: %v", homeworkType)
	}

	return metadata, hiddenMetadata, nil
}

type QuizAnswerType string

const (
	QuizAnswerTypeSingle QuizAnswerType = "single_answer"
	QuizAnswerTypeMulti  QuizAnswerType = "multi_answer"
	QuizAnswerTypeEmpty  QuizAnswerType = "empty"
)

type QuizMetadata struct {
	Questions []struct {
		AnswerType QuizAnswerType `json:"answer_type"`
		Question   string         `json:"question"`
		Options    []string       `json:"options"`
	} `json:"questions"`
}

type QuizHiddenMetadata struct {
	Answers [][]bool `json:"answers,omitempty"`
}

type OpenQuestionMetadata struct {
	Question        string `json:"question"`
	AllowFileAnswer bool   `json:"allow_file_answer"`
}

func (c *QuizMetadata) ToQuizResults(
	hiddenMetadata *QuizHiddenMetadata,
	submition [][]bool,
) ([]QuizResult, int64, error) {

	correctAnswers := hiddenMetadata.Answers

	if len(c.Questions) == 0 || len(c.Questions) != len(correctAnswers) {
		return nil, 0, fmt.Errorf("invalid quiz")
	}
	if len(submition) != len(c.Questions) {
		return nil, 0, fmt.Errorf("invalid submition")
	}

	quizResults := []QuizResult{}
	score := int64(0)

	totalQuestions := int64(0)
	for i := range len(correctAnswers) {
		switch c.Questions[i].AnswerType {
		case QuizAnswerTypeEmpty:
			continue
		case QuizAnswerTypeSingle, QuizAnswerTypeMulti:
			totalQuestions++
		default:
			return nil, 0, fmt.Errorf("wrong quiz type: %v", c.Questions[i].AnswerType)
		}

		if len(c.Questions[i].Options) == 0 || len(c.Questions[i].Options) != len(correctAnswers[i]) {
			return nil, 0, fmt.Errorf("invalid quiz question")
		}
		if len(submition[i]) != len(c.Questions[i].Options) {
			return nil, 0, fmt.Errorf("wrong number of answers in submition")
		}

		numOfCorrect := int64(0)
		for j := range c.Questions[i].Options {
			if submition[i][j] == correctAnswers[i][j] {
				numOfCorrect++
			}
		}
		if int(numOfCorrect) == len(c.Questions[i].Options) {
			score += MaxScore
		}

		quizResults = append(quizResults, QuizResult{
			UserAnswers:    submition[i],
			CorrectAnswers: correctAnswers[i],
		})
	}

	score = score / totalQuestions

	return quizResults, score, nil
}

func (c *QuizMetadata) ToQuestionResult(
	hiddenMetadata *QuizHiddenMetadata,
	questionIndex int,
	questionAnswer []bool,
) (*QuizResult, error) {

	if len(c.Questions) == 0 {
		return nil, fmt.Errorf("quiz not include any questions")
	}

	if questionIndex < 0 || len(c.Questions) <= questionIndex {
		return nil, fmt.Errorf("question index outside the quiz scope")
	}

	question := c.Questions[questionIndex]
	correctAnswers := hiddenMetadata.Answers[questionIndex]

	switch question.AnswerType {
	case QuizAnswerTypeEmpty:
		return &QuizResult{}, nil
	case QuizAnswerTypeSingle, QuizAnswerTypeMulti:
	default:
		return nil, fmt.Errorf("unsupported quiz type: %v", question.AnswerType)
	}

	if len(question.Options) == 0 || len(question.Options) != len(correctAnswers) {
		return nil, fmt.Errorf("invalid quiz question")
	}
	if len(questionAnswer) != len(question.Options) {
		return nil, fmt.Errorf("wrong number of answers in submition")
	}

	return &QuizResult{
		UserAnswers:    questionAnswer,
		CorrectAnswers: correctAnswers,
	}, nil
}

type MuxVideoMetadata struct {
	AssetID    string `json:"asset_id"`
	PlaybackID string `json:"playback_id"`
}
