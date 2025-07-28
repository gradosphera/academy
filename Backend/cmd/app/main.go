package main

import (
	"academy/internal/api/server"
	v1 "academy/internal/api/v1"
	"academy/internal/config"
	"academy/internal/cron"
	"academy/internal/database"
	"academy/internal/logger"
	"academy/internal/service"
	"academy/internal/service/jwt"
	"academy/internal/storage/cache"
	"academy/internal/storage/repository"
	"log"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}

	Build().Run()
}

func Build() *fx.App {
	return fx.New(
		fx.Options(
			config.Module,
			logger.Module,
		),
		database.Module(),
		jwt.Module(),

		repository.Module(),
		cache.Module(),

		service.Module(),
		server.Module(),

		v1.Module(),
		cron.Module(),
	)
}
