package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ReviewService struct {
	reviewRepository   *repository.ReviewRepository
	transactionManager *repo.TransactionManager
}

func NewReviewService(
	reviewRepository *repository.ReviewRepository,
	transactionManager *repo.TransactionManager,
) *ReviewService {

	return &ReviewService{
		reviewRepository:   reviewRepository,
		transactionManager: transactionManager,
	}
}

func (s *ReviewService) Create(ctx context.Context, review *model.Review) error {
	err := s.reviewRepository.Create(ctx, review)
	if err != nil {
		return fmt.Errorf("failed to create a review: %w", err)
	}

	return nil
}

func (s *ReviewService) GetByID(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	review, err := s.reviewRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get review by id: %w", err)
	}

	return review, nil
}

func (s *ReviewService) GetByUser(
	ctx context.Context, userID, productID uuid.UUID,
) ([]*model.Review, error) {

	reviews, err := s.reviewRepository.GetByUser(ctx, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get review by user: %w", err)
	}

	return reviews, nil
}

func (s *ReviewService) Update(ctx context.Context, review *model.Review) error {
	err := s.reviewRepository.Update(ctx, review)
	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	return nil
}

func (s *ReviewService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.reviewRepository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete review by id: %w", err)
	}

	return nil
}
