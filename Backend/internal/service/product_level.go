package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProductLevelService struct {
	productLevelRepository *repository.ProductLevelRepository
	productRepository      *repository.ProductRepository
	paymentRepository      *repository.PaymentRepository
	transactionManager     *repo.TransactionManager
}

func NewProductLevelService(
	productLevelRepository *repository.ProductLevelRepository,
	productRepository *repository.ProductRepository,
	paymentRepository *repository.PaymentRepository,
	transactionManager *repo.TransactionManager,
) *ProductLevelService {

	return &ProductLevelService{
		productLevelRepository: productLevelRepository,
		productRepository:      productRepository,
		paymentRepository:      paymentRepository,
		transactionManager:     transactionManager,
	}
}

func (s *ProductLevelService) Create(
	ctx context.Context,
	productLevel *model.ProductLevel,
	lessonIDs []uuid.UUID,
) error {

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.productLevelRepository.WithTx(tx).Create(ctx, productLevel)
		if err != nil {
			return fmt.Errorf("failed to create a product level: %w", err)
		}

		err = s.productLevelRepository.WithTx(tx).AddLessonRelations(ctx, productLevel.ID, lessonIDs...)
		if err != nil {
			return fmt.Errorf("failed to create product level relations: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create a product level: %w", err)
	}

	return nil
}

func (s *ProductLevelService) Update(
	ctx context.Context,
	productLevel *model.ProductLevel,
	addLessons []uuid.UUID,
	removeLessons []uuid.UUID,
) error {

	err := s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.productLevelRepository.WithTx(tx).Update(ctx, productLevel)
		if err != nil {
			return fmt.Errorf("failed to update product level: %w", err)
		}

		if len(addLessons) != 0 {
			err := s.productLevelRepository.WithTx(tx).AddLessonRelations(ctx, productLevel.ID, addLessons...)
			if err != nil {
				return fmt.Errorf("failed to create product level relations: %w", err)
			}
		}

		if len(removeLessons) != 0 {
			err := s.productLevelRepository.WithTx(tx).RemoveLessonRelations(ctx, productLevel.ID, removeLessons...)
			if err != nil {
				return fmt.Errorf("failed to remove product level relations: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *ProductLevelService) Delete(
	ctx context.Context,
	productLevelID uuid.UUID,
) error {

	err := s.productLevelRepository.Delete(ctx, productLevelID)
	if err != nil {
		return fmt.Errorf("failed to delete a product level: %w", err)
	}

	return nil
}

func (s *ProductLevelService) GetByID(ctx context.Context, id uuid.UUID) (*model.ProductLevel, error) {
	levels, err := s.productLevelRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product level by id: %w", err)
	}

	if len(levels) == 0 {
		return nil, fmt.Errorf("product level not found")
	}

	return levels[0], nil
}

func (s *ProductLevelService) IsProductLevelUnlocked(
	ctx context.Context,
	productLevelID, userID uuid.UUID,
) (bool, error) {

	isUnlocked, err := s.productLevelRepository.IsProductLevelUnlocked(ctx, productLevelID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check product level access: %w", err)
	}

	return isUnlocked, nil
}

func (s *ProductLevelService) GetProducts(ctx context.Context, id []uuid.UUID) ([]*model.Product, error) {
	products, err := s.productLevelRepository.GetProducts(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product level by id: %w", err)
	}

	return products, nil
}

func (s *ProductLevelService) CreateProductLevelInvite(ctx context.Context, invite *model.ProductLevelInvite) error {
	err := s.productLevelRepository.CreateProductLevelInvite(ctx, invite)
	if err != nil {
		return fmt.Errorf("failed to create product level invite: %w", err)
	}

	return nil
}

func (s *ProductLevelService) ProductLevelInvites(
	ctx context.Context,
	productID uuid.UUID,
	filter *model.FilterProductLevelInvitesRequest,
) ([]*model.ProductLevelInvite, int, error) {

	invites, total, err := s.productLevelRepository.ProductLevelInvites(ctx, productID, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get product level invites: %w", err)
	}

	return invites, total, nil
}

func (s *ProductLevelService) ClaimInvite(
	ctx context.Context,
	inviteID, userID uuid.UUID,
) (*model.ProductLevel, error) {

	invites, err := s.productLevelRepository.FindProductLevelInvites(ctx, inviteID, uuid.Nil, uuid.Nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find product level invite by id: %w", err)
	}

	if len(invites) != 1 {
		return nil, fmt.Errorf("invite is not found")
	}

	invite := invites[0]

	if invite.UserID != uuid.Nil && invite.UserID != userID {
		return nil, fmt.Errorf("invite already claimed")
	}

	productLevel, err := s.GetByID(ctx, invite.ProductLevelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product level by id: %w", err)
	}

	product, err := s.productRepository.GetByID(ctx, productLevel.ProductID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}

	invite.UserID = userID
	invite.UpdatedAt = time.Now()

	err = s.transactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		if err := s.productLevelRepository.WithTx(tx).UpdateInvite(ctx, invite); err != nil {
			return fmt.Errorf("failed to update product level invite: %w", err)
		}

		payment := model.NewFreePaymentForProductLevel(userID, product, productLevel)

		if err := s.paymentRepository.WithTx(tx).Create(ctx, payment); err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return productLevel, nil
}
