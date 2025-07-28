package database

import (
	"academy/internal/config"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func CreateSQLConn(cfg *config.Config) (*bun.DB, error) {
	conf, err := pgxpool.ParseConfig(cfg.PG.GetConnectString())
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	db := bun.NewDB(sqlDB, pgdialect.New())

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
