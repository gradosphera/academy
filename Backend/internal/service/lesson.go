package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type LessonService struct {
	lessonRepository       *repository.LessonRepository
	productRepository      *repository.ProductRepository
	productLevelRepository *repository.ProductLevelRepository
	transactionManager     *repo.TransactionManager
}

func NewLessonService(
	lessonRepository *repository.LessonRepository,
	productRepository *repository.ProductRepository,
	productLevelRepository *repository.ProductLevelRepository,
	transactionManager *repo.TransactionManager,
) *LessonService {

	return &LessonService{
		lessonRepository:       lessonRepository,
		productRepository:      productRepository,
		productLevelRepository: productLevelRepository,
		transactionManager:     transactionManager,
	}
}

func (s *LessonService) Create(
	ctx context.Context,
	newLesson *model.Lesson,
	newLessonIndex *int64,
	product *model.Product,
	productLevelIDs []uuid.UUID,
) error {

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		if 0 < len(product.Lessons) {
			if newLessonIndex != nil && 0 <= int(*newLessonIndex) && int(*newLessonIndex) < len(product.Lessons) {
				newLesson.Index = *newLessonIndex

				if product.LessonAccess == model.LessonAccessSequential && *newLessonIndex != 0 {
					newLesson.PreviousLessonID = product.Lessons[*newLessonIndex-1].ID
				}

				prevLessonID := uuid.Nil
				for i := *newLessonIndex; int(i) < len(product.Lessons); i++ {
					product.Lessons[i].Index += 1

					if product.LessonAccess == model.LessonAccessSequential {
						product.Lessons[i].PreviousLessonID = prevLessonID
					}

					err := s.lessonRepository.WithTx(tx).Update(ctx, product.Lessons[i])
					if err != nil {
						return fmt.Errorf("failed to update lesson: %w", err)
					}

					prevLessonID = product.Lessons[i].ID
				}
			} else {
				newLesson.Index = int64(len(product.Lessons))

				if product.LessonAccess == model.LessonAccessSequential {
					newLesson.PreviousLessonID = product.Lessons[len(product.Lessons)-1].ID
				}
			}
		}

		if err := s.lessonRepository.WithTx(tx).Create(ctx, newLesson); err != nil {
			return fmt.Errorf("failed to create a lesson: %w", err)
		}

		// Update PreviousLessonID for next lesson after newLesson only after saving newLesson.
		if product.LessonAccess == model.LessonAccessSequential &&
			newLessonIndex != nil &&
			0 <= int(*newLessonIndex) && int(*newLessonIndex) < len(product.Lessons) {

			nextLesson := product.Lessons[*newLessonIndex]
			nextLesson.PreviousLessonID = newLesson.ID

			if err := s.lessonRepository.WithTx(tx).Update(ctx, nextLesson); err != nil {
				return fmt.Errorf("failed to update new to the new lesson: %w", err)
			}
		}

		for _, productLevelID := range productLevelIDs {
			err := s.productLevelRepository.WithTx(tx).AddLessonRelations(ctx, productLevelID, newLesson.ID)
			if err != nil {
				return fmt.Errorf("failed to add lesson relation: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *LessonService) GetByID(ctx context.Context, id, userID uuid.UUID) (*model.Lesson, error) {
	lesson, err := s.lessonRepository.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson by id: %w", err)
	}

	return lesson, nil
}

func (s *LessonService) Update(ctx context.Context, lesson *model.Lesson) error {
	err := s.lessonRepository.Update(ctx, lesson)
	if err != nil {
		return fmt.Errorf("failed to update lesson: %w", err)
	}

	return nil
}

func (s *LessonService) IsLessonUnlocked(ctx context.Context, lessonID, userID uuid.UUID) (bool, error) {
	ok, err := s.lessonRepository.IsLessonUnlocked(ctx, lessonID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check lesson access: %w", err)
	}

	return ok, nil
}

func (s *LessonService) UnlockedLessons(ctx context.Context, productID, userID uuid.UUID) ([]model.UnlockedLesson, error) {
	lessons, err := s.lessonRepository.UnlockedLessons(ctx, productID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check unlocked lessons: %w", err)
	}

	return lessons, nil
}

func (s *LessonService) Delete(ctx context.Context, id uuid.UUID, product *model.Product) error {
	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.lessonRepository.WithTx(tx).Delete(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete lesson by id: %w", err)
		}

		if product.LessonAccess != model.LessonAccessSequential {
			return nil
		}

		prevLessonID := uuid.Nil
		lessonIndex := int64(0)
		for _, lesson := range product.Lessons {
			if lesson.ID == id {
				continue
			}

			if lesson.PreviousLessonID != prevLessonID || lesson.Index != lessonIndex {

				lesson.PreviousLessonID = prevLessonID
				lesson.Index = lessonIndex

				err := s.lessonRepository.WithTx(tx).Update(ctx, lesson)
				if err != nil {
					return fmt.Errorf("failed to set prev lesson: %w", err)
				}
			}

			prevLessonID = lesson.ID
			lessonIndex++
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
