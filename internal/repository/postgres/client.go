package postgres

import (
	"ServiceManager/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPGPool(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, ok := ctx.Value("config").(*config.Config)
	if !ok {
		return nil, fmt.Errorf("config is not found in context")
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost,
		cfg.DBPort, cfg.DBName)

	pool, errPGX := pgxpool.New(ctx, connStr)
	if errPGX != nil {
		return nil, errPGX
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
