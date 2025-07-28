package repository

import (
	"academy/internal/database/repository"
	"academy/internal/model"

	"github.com/google/uuid"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			repository.NewGenericRepository[model.User, uuid.UUID],
			NewUserRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.MiniApp, uuid.UUID],
			NewMiniAppRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Product, uuid.UUID],
			NewProductRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Lesson, uuid.UUID],
			NewLessonRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Material, uuid.UUID],
			NewMaterialRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.LessonProgress, uuid.UUID],
			NewLessonProgressRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.ProductLevel, uuid.UUID],
			NewProductLevelRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Payment, uuid.UUID],
			NewPaymentRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Review, uuid.UUID],
			NewReviewRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.Chunk, uuid.UUID],
			NewChunkRepository,
		),
		fx.Provide(
			repository.NewGenericRepository[model.JettonTransfer, uuid.UUID],
			NewJettonTransferRepository,
		),
	)
}
