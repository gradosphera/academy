package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"academy/internal/types"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/xuri/excelize/v2"
)

var ErrNoData = errors.New("no data")

type ProductService struct {
	productRepository      *repository.ProductRepository
	productLevelRepository *repository.ProductLevelRepository
	lessonRepository       *repository.LessonRepository
	paymentRepository      *repository.PaymentRepository
	transactionManager     *repo.TransactionManager
}

func NewProductService(
	productRepository *repository.ProductRepository,
	productLevelRepository *repository.ProductLevelRepository,
	lessonRepository *repository.LessonRepository,
	paymentRepository *repository.PaymentRepository,
	transactionManager *repo.TransactionManager,
) *ProductService {

	return &ProductService{
		productRepository:      productRepository,
		productLevelRepository: productLevelRepository,
		lessonRepository:       lessonRepository,
		paymentRepository:      paymentRepository,
		transactionManager:     transactionManager,
	}
}

func (s *ProductService) Create(ctx context.Context, product *model.Product) error {
	err := s.productRepository.Create(ctx, product)
	if err != nil {
		return fmt.Errorf("failed to create a product: %w", err)
	}

	return nil
}

func (s *ProductService) GetByID(ctx context.Context, id uuid.UUID, includeRelations bool) (*model.Product, error) {
	product, err := s.productRepository.GetByID(ctx, id, includeRelations)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}

	return product, nil
}

func (s *ProductService) Update(
	ctx context.Context,
	product *model.Product,
	applyNewLessonAccess *model.LessonAccess,
	applyNewReleaseDate *types.Time,
	applyNewAccessTime *types.Interval,
) error {

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.productRepository.WithTx(tx).Update(ctx, product)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		if applyNewLessonAccess != nil {
			prevLessonID := uuid.Nil
			for _, lesson := range product.Lessons {
				lesson.PreviousLessonID = prevLessonID

				if *applyNewLessonAccess != model.LessonAccessScheduled {
					lesson.ReleaseDate = types.Time{}
				}

				err := s.lessonRepository.WithTx(tx).Update(ctx, lesson)
				if err != nil {
					return fmt.Errorf("failed to set prev lesson: %w", err)
				}

				if *applyNewLessonAccess == model.LessonAccessSequential {
					prevLessonID = lesson.ID
				}
			}
		}

		if applyNewReleaseDate != nil {
			err := s.paymentRepository.WithTx(tx).UpdateAccessStart(ctx, product.ID, *applyNewReleaseDate)
			if err != nil {
				return fmt.Errorf("failed to update payments access_start: %w", err)
			}
		}

		if applyNewAccessTime != nil {
			err := s.productLevelRepository.WithTx(tx).UpdateDuration(ctx, product.ID, *applyNewAccessTime)
			if err != nil {
				return fmt.Errorf("failed to update product levels duration: %w", err)
			}
			for _, l := range product.Levels {
				l.Duration = *applyNewAccessTime
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *ProductService) ReorderLessons(
	ctx context.Context,
	productID uuid.UUID,
	req *model.ReorderProductLessonsRequest,
) (*model.Product, error) {

	product, err := s.productRepository.GetByID(ctx, productID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}

	productLessons := make(map[uuid.UUID]*model.Lesson)
	for _, lesson := range product.Lessons {
		productLessons[lesson.ID] = lesson
	}

	now := time.Now().UTC()
	lessonIndex := int64(0)
	var prevLessonID uuid.UUID
	for _, module := range req.Modules {
		for _, lessonID := range module.LessonIDs {
			lesson, ok := productLessons[lessonID]
			if !ok {
				return nil, fmt.Errorf("lesson is not included in the product: %v", lessonID)
			}

			if product.LessonAccess == model.LessonAccessSequential {
				lesson.PreviousLessonID = prevLessonID
			}

			lesson.ModuleName = module.ModuleName
			lesson.Index = lessonIndex
			lesson.UpdatedAt = now

			lessonIndex++
			prevLessonID = lesson.ID
		}
	}

	if int(lessonIndex)+len(req.LessonsToDelete) != len(product.Lessons) {
		return nil, fmt.Errorf("lessons included in request do not include all product lessons")
	}

	newProductLessons := make([]*model.Lesson, 0, len(product.Lessons))
	err = s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		for _, lessonID := range req.LessonsToDelete {
			if _, ok := productLessons[lessonID]; !ok {
				return fmt.Errorf("lesson to delete is not included in the product: %v", lessonID)
			}

			if err := s.lessonRepository.WithTx(tx).Delete(ctx, lessonID); err != nil {
				return fmt.Errorf("failed to delete lesson: %w", err)
			}

			delete(productLessons, lessonID)
		}

		for _, l := range productLessons {
			newProductLessons = append(newProductLessons, l)
			if err := s.lessonRepository.WithTx(tx).Update(ctx, l); err != nil {
				return fmt.Errorf("failed to update lesson: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	slices.SortFunc(newProductLessons, func(a, b *model.Lesson) int {
		return int(a.Index - b.Index)
	})

	product.Lessons = newProductLessons

	return product, nil
}

func (s *ProductService) ReorderLevels(
	ctx context.Context,
	productID uuid.UUID,
	req *model.ReorderProductLevelsRequest,
) (*model.Product, error) {

	product, err := s.productRepository.GetByID(ctx, productID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}

	productLevels := make(map[uuid.UUID]*model.ProductLevel)
	for _, productLevel := range product.Levels {
		productLevels[productLevel.ID] = productLevel
	}

	now := time.Now().UTC()
	productLevelIndex := int64(0)
	for _, productLevelID := range req.ProductLevels {
		productLevel, ok := productLevels[productLevelID]
		if !ok {
			return nil, fmt.Errorf("product level is not included in the product: %v", productLevelID)
		}

		productLevel.Index = productLevelIndex
		productLevel.UpdatedAt = now

		productLevelIndex++
	}

	if int(productLevelIndex) != len(productLevels) {
		return nil, fmt.Errorf("product levels included in request do not include all product levels")
	}

	newProductLevels := make([]*model.ProductLevel, 0, len(product.Levels))
	err = s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		for _, l := range productLevels {
			newProductLevels = append(newProductLevels, l)
			if err := s.productLevelRepository.WithTx(tx).Update(ctx, l); err != nil {
				return fmt.Errorf("failed to update product level: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	slices.SortFunc(newProductLevels, func(a, b *model.ProductLevel) int {
		return int(a.Index - b.Index)
	})

	product.Levels = newProductLevels

	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.productRepository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product by id: %w", err)
	}

	return nil
}

func (s *ProductService) Feedback(ctx context.Context, productID uuid.UUID) (*model.ProductFeedback, error) {
	feedback, err := s.productRepository.Feedback(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to product feedback: %w", err)
	}

	return feedback, nil
}

func (s *ProductService) Students(
	ctx context.Context, productID uuid.UUID, usernameSearch string,
	limit, offset uint,
) (*model.ProductStudents, error) {

	students, err := s.productRepository.Students(ctx, productID, usernameSearch, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to product students: %w", err)
	}

	return students, nil
}

func (s *ProductService) StudentsExportToExcel(
	ctx context.Context,
	product *model.Product,
	req *model.ExportProductStudentsRequest,
) ([]byte, error) {

	dateFrom, err := time.Parse(time.DateOnly, req.DateFrom)
	if err != nil {
		return nil, fmt.Errorf("error parsing DateFrom: %w", err)
	}
	dateTo, err := time.Parse(time.DateOnly, req.DateTo)
	if err != nil {
		return nil, fmt.Errorf("error parsing DateTo: %w", err)
	}

	students, lessonNames, err := s.productRepository.StudentsDetails(ctx, product, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("failed to get product students details: %w", err)
	}

	if len(students) == 0 {
		return nil, ErrNoData
	}

	f := excelize.NewFile()

	headers := []string{
		"User ID", "First Name", "Last Name",
		"Telegram ID", "Telegram Username",
		"Completed Lessons", "Total Lessons",
	}
	headers = append(headers, lessonNames...)

	cell, _ := excelize.CoordinatesToCellName(1, 1)
	err = f.SetSheetRow("Sheet1", cell, &headers)
	if err != nil {
		return nil, fmt.Errorf("failed to set headers: %w", err)
	}
	for i, student := range students {
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		studentRow := []any{
			student.UserID.String(),
			student.FirstName,
			student.LastName,
			student.TelegramID,
			student.TelegramUsername,
			student.CompletedLessons,
			student.TotalLessons,
		}
		for _, l := range student.Lessons {
			studentRow = append(studentRow, l)
		}
		err = f.SetSheetRow("Sheet1", cell, &studentRow)
		if err != nil {
			return nil, fmt.Errorf("failed to set student#%d: %w", i, err)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("could not write excel to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

// CheckProductAccess creates new product access record. If exists then just
// updates updated_at field.
func (s *ProductService) CheckProductAccess(
	ctx context.Context,
	productAccess *model.ProductAccess,
) (*model.ProductAccess, error) {

	productAccess, err := s.productRepository.CheckProductAccess(ctx, productAccess)
	if err != nil {
		return nil, fmt.Errorf("failed to check product access: %w", err)
	}

	return productAccess, nil
}

func (s *ProductService) ProductAccessByUser(
	ctx context.Context,
	miniAppID, userID uuid.UUID,
) ([]*model.ProductAccess, error) {

	accesses, err := s.productRepository.ProductAccessByUser(ctx, miniAppID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product accesses: %w", err)
	}

	return accesses, nil
}
