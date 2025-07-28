package cache

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("cache",
		fx.Provide(
			NewJWTCacheStorage,
		),
	)
}
