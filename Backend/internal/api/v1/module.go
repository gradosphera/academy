package v1

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("v1",
		fx.Provide(
			NewV1Handler,
		),
		fx.Invoke(func(app *fiber.App, h *V1Handler) {
			h.RegisterRoutes(app)
		}),
	)
}
