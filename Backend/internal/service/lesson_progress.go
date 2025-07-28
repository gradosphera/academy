package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LessonProgressService struct {
	lessonProgressRepository *repository.LessonProgressRepository
	transactionManager       *repo.TransactionManager
}

func NewLessonProgressService(
	lessonProgressRepository *repository.LessonProgressRepository,
	transactionManager *repo.TransactionManager,
) *LessonProgressService {

	return &LessonProgressService{
		lessonProgressRepository: lessonProgressRepository,
		transactionManager:       transactionManager,
	}
}

func (s *LessonProgressService) CreateOrUpdate(
	ctx context.Context,
	lessonProgress *model.LessonProgress,
) error {

	err := s.lessonProgressRepository.CreateOrUpdate(ctx, lessonProgress)
	if err != nil {
		return fmt.Errorf("failed to create a lesson_progress: %w", err)
	}

	return nil
}

func (s *LessonProgressService) GetByID(
	ctx context.Context,
	userID, lessonID uuid.UUID,
) (*model.LessonProgress, error) {

	lessonProgress, err := s.lessonProgressRepository.GetByID(ctx, userID, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson_progress by id: %w", err)
	}

	return lessonProgress, nil
}

func (s *LessonProgressService) GetByProductID(
	ctx context.Context,
	userID, productID uuid.UUID,
) ([]*model.LessonProgress, error) {

	progress, err := s.lessonProgressRepository.GetByProductID(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson_progress by product: %w", err)
	}

	return progress, nil
}

func (s *LessonProgressService) ProductHomework(
	ctx context.Context,
	productID uuid.UUID,
	filter *model.FilterProductHomeworkRequest,
) ([]*model.LessonProgress, int, error) {

	progressFilter := &model.FilterLessonProgressRequest{
		ProductID: []uuid.UUID{productID},
		UserID:    filter.UserID,
		LessonID:  filter.LessonID,
		Status:    []model.LessonProgressStatus{model.LessonProgressStatusPending},

		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	lessonProgress, total, err := s.lessonProgressRepository.Find(ctx, progressFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find lesson progress: %w", err)
	}

	return lessonProgress, total, nil
}

func (s *LessonProgressService) UserHomework(
	ctx context.Context,
	userID uuid.UUID,
	filter *model.FilterUserHomeworkRequest,
) ([]*model.LessonProgress, int, error) {

	progressFilter := &model.FilterLessonProgressRequest{
		UserID:   []uuid.UUID{userID},
		LessonID: filter.LessonID,
		Status:   filter.Status,

		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	lessonProgress, total, err := s.lessonProgressRepository.Find(ctx, progressFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find lesson progress: %w", err)
	}

	return lessonProgress, total, nil
}

func (s *LessonProgressService) FeedbackHomework(
	ctx context.Context,
	req *model.FeedbackHomeworkRequest,
) error {

	if req.Score < 0 || model.MaxScore < req.Score {
		return fmt.Errorf("invalid score: %v", req.Score)
	}

	progress, err := s.lessonProgressRepository.GetByID(ctx, req.UserID, req.LessonID)
	if err != nil {
		return fmt.Errorf("failed to get lesson progress: %w", err)
	}

	var progressData model.LessonProgressData
	err = json.Unmarshal(progress.Data, &progressData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal progress data: %w", err)
	}

	progressData.Feedback = req.Feedback

	newData, err := json.Marshal(progressData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal progress data: %w", err)
	}

	progress.Data = newData
	progress.Score = req.Score
	progress.Status = req.NewStatus
	progress.UpdatedAt = time.Now().UTC()

	err = s.lessonProgressRepository.Update(ctx, progress)
	if err != nil {
		return fmt.Errorf("failed to find lesson progress: %w", err)
	}

	return nil
}
