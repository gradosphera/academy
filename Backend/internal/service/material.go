package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type MaterialService struct {
	materialRepository *repository.MaterialRepository
	transactionManager *repo.TransactionManager
}

func NewMaterialService(
	materialRepository *repository.MaterialRepository,
	transactionManager *repo.TransactionManager,
) *MaterialService {

	return &MaterialService{
		materialRepository: materialRepository,
		transactionManager: transactionManager,
	}
}

func (s *MaterialService) Create(ctx context.Context, material *model.Material) error {
	err := s.materialRepository.Create(ctx, material)
	if err != nil {
		return fmt.Errorf("failed to create a material: %w", err)
	}

	return nil
}

func (s *MaterialService) GetByID(ctx context.Context, id uuid.UUID) (*model.Material, error) {
	material, err := s.materialRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get material by id: %w", err)
	}

	return material, nil
}

func (s *MaterialService) GetByFilename(ctx context.Context, filename string) (*model.Material, error) {
	material, err := s.materialRepository.GetByFilename(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get material by filename: %w", err)
	}

	return material, nil
}

func (s *MaterialService) FindPendingCompressing(ctx context.Context, limit int) ([]*model.Material, error) {
	material, err := s.materialRepository.FindByStatus(
		ctx, model.MaterialStatusPendingCompressing, false, limit, 0)

	if err != nil {
		return nil, fmt.Errorf("failed to find pending compressing materials: %w", err)
	}

	return material, nil
}

func (s *MaterialService) FindPendingMoveToMux(
	ctx context.Context, withMetadata bool, limit int,
) ([]*model.Material, error) {

	material, err := s.materialRepository.FindByStatus(
		ctx, model.MaterialStatusPendingMoveToMux, withMetadata, limit, 0)

	if err != nil {
		return nil, fmt.Errorf("failed to find pending moving to mux materials: %w", err)
	}

	return material, nil
}

func (s *MaterialService) FindMuxAssetsToDelete(ctx context.Context, limit int) ([]string, error) {
	assets, err := s.materialRepository.FindMuxAssetsToDelete(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find mux assets to delete: %w", err)
	}

	return assets, nil
}

func (s *MaterialService) FullyDeleleMuxAsset(ctx context.Context, assetID string) error {
	err := s.materialRepository.FullyDeleleMuxAsset(ctx, assetID)
	if err != nil {
		return fmt.Errorf("failed to find mux assets to delete: %w", err)
	}

	return nil
}

func (s *MaterialService) Update(ctx context.Context, material *model.Material) error {
	err := s.materialRepository.Update(ctx, material)
	if err != nil {
		return fmt.Errorf("failed to update material: %w", err)
	}

	return nil
}

func (s *MaterialService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.materialRepository.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete material by id: %w", err)
	}

	return nil
}
