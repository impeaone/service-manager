package repository

import (
	"ServiceManager/internal/domain"
	"context"
	"fmt"
)

var (
	ServiceNotFoundError = fmt.Errorf("service not found")
)

type ServiceRepository interface {
	Create(cxt context.Context, userID string, service *domain.Service) error
	GetByID(ctx context.Context, userID, serviceID string) (*domain.Service, error)
	GetAll(ctx context.Context, userID string) ([]*domain.Service, error)
	Update(ctx context.Context, userID string, service *domain.Service) error
	Delete(ctx context.Context, userID string, serviceID string) error

	IncrementWebHookExecutions(ctx context.Context, serviceID, hookID string) error
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	VerifyUser(ctx context.Context, email string) error
}
