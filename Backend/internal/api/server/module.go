package server

import (
	"academy/internal/config"
	"context"
	"log"
	"net"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("server",
		fx.Provide(
			NewServer,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, app *fiber.App, cfg *config.Config) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							addr := net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port)
							if err := app.Listen(addr); err != nil {
								log.Println("failed to start handler listening", err)
							}
						}()

						return nil
					},
					OnStop: func(ctx context.Context) error {
						return app.ShutdownWithContext(ctx)
					},
				},
				)
			},
		),
	)
}
