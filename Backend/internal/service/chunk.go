package service

import (
	repo "academy/internal/database/repository"
	"academy/internal/model"
	"academy/internal/storage/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ChunkService struct {
	chunkRepository    *repository.ChunkRepository
	transactionManager *repo.TransactionManager
}

func NewChunkService(
	chunkRepository *repository.ChunkRepository,
	transactionManager *repo.TransactionManager,
) *ChunkService {

	return &ChunkService{
		chunkRepository:    chunkRepository,
		transactionManager: transactionManager,
	}
}

func (s *ChunkService) Create(ctx context.Context, chunk *model.Chunk) error {
	err := s.chunkRepository.Create(ctx, chunk)
	if err != nil {
		return fmt.Errorf("failed to create a chunk: %w", err)
	}

	return nil
}

func (s *ChunkService) GetByID(ctx context.Context, materialID uuid.UUID, index ...int64) ([]*model.Chunk, error) {
	chunks, err := s.chunkRepository.GetByID(ctx, materialID, index...)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunks by id: %w", err)
	}

	return chunks, nil
}

func (s *ChunkService) Delete(ctx context.Context, materialID uuid.UUID) error {
	err := s.chunkRepository.Delete(ctx, materialID)
	if err != nil {
		return fmt.Errorf("failed to delete chunks by id: %w", err)
	}

	return nil
}
