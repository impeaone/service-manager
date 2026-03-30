package postgres

import (
	"ServiceManager/internal/domain"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, email, name, is_verified)
        VALUES ($1, $2, $3, $4)
    `

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	_, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.Name, user.IsVerified)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT id, email, name, is_verified, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
        SELECT id, email, name, is_verified, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	var user domain.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) VerifyUser(ctx context.Context, email string) error {
	query := `
        UPDATE users 
        SET is_verified = TRUE, updated_at = NOW()
        WHERE email = $1
    `

	cmdTag, err := r.pool.Exec(ctx, query, email)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
