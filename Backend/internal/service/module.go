package service

import (
	"academy/internal/service/security"
	"academy/internal/service/telegram"
	"academy/internal/service/ton"
	"academy/internal/service/upload"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("service",
		fx.Provide(
			NewJWTService,

			NewMiniAppService,
			NewUserService,
			NewProductService,
			NewLessonService,
			NewMaterialService,
			NewChunkService,

			NewLessonProgressService,
			NewProductLevelService,

			NewPaymentService,
			NewReviewService,

			ton.NewService,
			upload.NewService,
			telegram.NewService,
			security.NewService,
		),
	)
}
