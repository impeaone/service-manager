package postgres

import (
	"ServiceManager/internal/config"
	"ServiceManager/internal/domain"
	"ServiceManager/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func InitPostgres(ctx context.Context) (*Postgres, error) {
	pool, err := initPGPool(ctx)
	if err != nil {
		return nil, err
	}
	return &Postgres{
		pool: pool,
	}, nil

}

func initPGPool(ctx context.Context) (*pgxpool.Pool, error) {
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

func (p *Postgres) Create(ctx context.Context, service *domain.Service) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// TODO:
	query := `SELECT ID FROM Services WHERE ID = $1;`
	_, err = p.pool.Exec(ctx, query, service.ID)
	if err != nil && errors.As(err, &pgx.ErrNoRows) {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	if errors.As(err, &pgx.ErrNoRows) {
		return fmt.Errorf("service already exists")
	}

	err = p.insertService(ctx, tx, service)
	if err != nil {
		return fmt.Errorf("failed to insert service: %w", err)
	}

	if len(service.WebHooks) > 0 {
		err = p.insertWebhooks(ctx, tx, service.ID, service.WebHooks)
		if err != nil {
			return fmt.Errorf("failed to insert webhooks: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *Postgres) insertService(ctx context.Context, tx pgx.Tx, service *domain.Service) error {
	query := `
        INSERT INTO services (id, name, status)
        VALUES ($1, $2, $3)
    `

	if service.ID == "" {
		service.ID = utils.GenerateUUID()
	}

	_, err := tx.Exec(ctx, query, service.ID, service.Name, service.Status)

	return err
}

func (p *Postgres) insertWebhooks(ctx context.Context, tx pgx.Tx, serviceID string, webhooks []domain.WebHook) error {
	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"webhooks"},
		[]string{"id", "service_id", "name", "path", "type", "method", "executions", "last_call"},
		pgx.CopyFromSlice(len(webhooks), func(i int) ([]any, error) {
			webhook := webhooks[i]

			if webhook.ID == "" {
				webhook.ID = utils.GenerateUUID()
			}

			return []any{
				webhook.ID,
				serviceID,
				webhook.Name,
				webhook.Path,
				webhook.Type,
				webhook.Method,
				webhook.Executions,
				webhook.LastCall,
			}, nil
		}),
	)

	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	if copyCount != int64(len(webhooks)) {
		return fmt.Errorf("expected to copy %d rows, copied %d", len(webhooks), copyCount)
	}

	return nil
}

func (p *Postgres) Update(service *domain.Service) error {
	return nil
}

func (p *Postgres) Delete(serviceID string) error {
	return nil
}

func (p *Postgres) IncrementWebHookExecutions(serviceID, hookID string) error {
	return nil
}

func (p *Postgres) GetAll() ([]*domain.Service, error) {
	return []*domain.Service{}, nil
}

func (p *Postgres) GetByID(serviceID string) (*domain.Service, error) {
	return &domain.Service{}, nil
}
