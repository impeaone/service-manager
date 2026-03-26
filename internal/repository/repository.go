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
	Create(cxt context.Context, service *domain.Service) error
	GetByID(serviceID string) (*domain.Service, error)
	GetAll() ([]*domain.Service, error)
	Update(service *domain.Service) error
	Delete(serviceID string) error

	IncrementWebHookExecutions(serviceID, hookID string) error
}
