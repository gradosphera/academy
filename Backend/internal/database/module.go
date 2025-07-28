package database

import (
	"academy/internal/database/repository"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("database",
		fx.Provide(
			CreateSQLConn,
			CreateRedisClient,
		),
		fx.Provide(
			repository.NewTransactionManager,
		),
		fx.Provide(
			fx.Annotate(
				repository.NewDBWrapper,
				fx.As(
					new(repository.DB),
				)),
		),
	)
}
