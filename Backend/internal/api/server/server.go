package server

import (
	"academy/internal/api/middleware"
	"academy/internal/config"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

const defaultBodyLimit = 100_000_000 // 100 MB

func NewServer(cfg *config.Config, logger *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Academy",
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		BodyLimit:    defaultBodyLimit,
	})

	corsConfig := cors.Config{
		AllowOrigins:     strings.Split(cfg.HTTP.AllowOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: cfg.HTTP.AllowCredentials,
		MaxAge:           int(12 * time.Hour),
	}

	app.Use(cors.New(corsConfig))

	logMiddleware := middleware.NewLoggingMiddleware(logger)
	logMiddleware.RegisterLogger(cfg, app)
	app.Use(recoverer.New())

	app.Get("/", healthcheck.NewHealthChecker(healthcheck.Config{
		Probe: func(c fiber.Ctx) bool {
			_, err := c.WriteString("server is running")
			return err == nil
		},
	}))

	return app
}
