package postgres

import (
	"ServiceManager/internal/domain"
	"ServiceManager/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceRepository struct {
	pool *pgxpool.Pool
}

func NewServiceRepository(pool *pgxpool.Pool) *ServiceRepository {
	return &ServiceRepository{
		pool: pool,
	}

}

func (p *ServiceRepository) Stop(ctx context.Context) error {
	endChan := make(chan error, 1)

	go func() {
		p.pool.Close()
		endChan <- nil
		close(endChan)
	}()

	select {
	case <-ctx.Done():
		go func() {
			p.pool.Reset()
			p.pool.Close()
		}()
		return ctx.Err()
	case err := <-endChan:
		return err
	}
}

func (p *ServiceRepository) Create(ctx context.Context, userID string, service *domain.Service) error {
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

	query := `SELECT ID FROM Services WHERE ID = $1;`
	_, err = p.pool.Exec(ctx, query, service.ID)
	if err != nil && errors.As(err, &pgx.ErrNoRows) {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	if errors.As(err, &pgx.ErrNoRows) {
		return fmt.Errorf("service already exists")
	}

	err = p.insertService(ctx, tx, userID, service)
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

func (p *ServiceRepository) insertService(ctx context.Context, tx pgx.Tx, userID string, service *domain.Service) error {
	query := `
        INSERT INTO services (id, name, status, user_id)
        VALUES ($1, $2, $3, $4)
    `

	if service.ID == "" {
		service.ID = utils.GenerateUUID()
	}

	_, err := tx.Exec(ctx, query, service.ID, service.Name, service.Status, userID)

	return err
}

func (p *ServiceRepository) insertWebhooks(ctx context.Context, tx pgx.Tx, serviceID string, webhooks []domain.WebHook) error {
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

func (p *ServiceRepository) Update(ctx context.Context, userID string, service *domain.Service) error {
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

	query := `
        UPDATE services 
        SET name = $3, status = $4, updated_at = NOW()
        WHERE id = $1 and user_id = $2;
    `

	cmdTag, err := tx.Exec(ctx, query, service.ID, userID, service.Name, service.Status)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("service with id %s not found", service.ID)
	}

	// 2. Обновляем веб-хуки: удаляем старые и вставляем новые
	// Сначала удаляем все существующие веб-хуки для этого сервиса
	deleteQuery := `DELETE FROM webhooks WHERE service_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, service.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing webhooks: %w", err)
	}

	// Затем вставляем новые веб-хуки, если они есть
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

func (p *ServiceRepository) Delete(ctx context.Context, userID, serviceID string) error {
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

	query := `DELETE FROM services WHERE id = $1 and user_id = $2;`
	cmdTag, err := tx.Exec(ctx, query, serviceID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("service with id %s not found", serviceID)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *ServiceRepository) IncrementWebHookExecutions(ctx context.Context, serviceID, hookID string) error {
	query := `
		update webhooks
        set executions = executions + 1,
            last_call = NOW()
		where id = $1 and service_id = $2
    `

	cmdTag, err := p.pool.Exec(ctx, query, hookID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to increment webhook executions: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("webhook with id %s and service_id %s not found", hookID, serviceID)
	}

	return nil
}

func (p *ServiceRepository) GetAll(ctx context.Context, userID string) ([]*domain.Service, error) {
	servicesQuery := `
        SELECT id, name, status, created_at, updated_at 
        FROM services 
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	rows, err := p.pool.Query(ctx, servicesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []*domain.Service
	serviceMap := make(map[string]*domain.Service)

	for rows.Next() {
		var service domain.Service
		err = rows.Scan(&service.ID, &service.Name, &service.Status, &service.CreatedAt, &service.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		service.WebHooks = []domain.WebHook{}
		services = append(services, &service)
		serviceMap[service.ID] = &service
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating services: %w", err)
	}

	if len(services) == 0 {
		return []*domain.Service{}, nil
	}

	webhooksQuery := `
        SELECT id, service_id, name, path, type, method, executions, last_call 
        FROM webhooks 
        WHERE service_id = ANY($1)
        ORDER BY name
    `

	serviceIDs := make([]string, 0, len(services))
	for _, s := range services {
		serviceIDs = append(serviceIDs, s.ID)
	}

	webhookRows, err := p.pool.Query(ctx, webhooksQuery, serviceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer webhookRows.Close()

	for webhookRows.Next() {
		var webhook domain.WebHook
		var serviceID string
		err = webhookRows.Scan(
			&webhook.ID,
			&serviceID,
			&webhook.Name,
			&webhook.Path,
			&webhook.Type,
			&webhook.Method,
			&webhook.Executions,
			&webhook.LastCall,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		if service, exists := serviceMap[serviceID]; exists {
			service.WebHooks = append(service.WebHooks, webhook)
		}
	}

	if err = webhookRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhooks: %w", err)
	}

	return services, nil
}

func (p *ServiceRepository) GetByID(ctx context.Context, userID, serviceID string) (*domain.Service, error) {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	serviceQuery := `
        SELECT id, name, status, created_at, updated_at 
        FROM services 
        WHERE id = $1
        AND user_id = $2
    `

	var service domain.Service
	err = tx.QueryRow(ctx, serviceQuery, serviceID, userID).Scan(
		&service.ID,
		&service.Name,
		&service.Status,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("service with id %s not found", serviceID)
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	webhooksQuery := `
        SELECT id, name, path, type, method, executions, last_call 
        FROM webhooks 
        WHERE service_id = $1
        ORDER BY name
    `

	rows, err := tx.Query(ctx, webhooksQuery, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	service.WebHooks = []domain.WebHook{}
	for rows.Next() {
		var webhook domain.WebHook
		err = rows.Scan(
			&webhook.ID,
			&webhook.Name,
			&webhook.Path,
			&webhook.Type,
			&webhook.Method,
			&webhook.Executions,
			&webhook.LastCall,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		service.WebHooks = append(service.WebHooks, webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhooks: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &service, nil
}
