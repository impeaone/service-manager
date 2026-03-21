package repository

import (
	"ServiceManager/internal/domain"
	"errors"
)

var (
	ServiceNotFoundError = errors.New("service not found")
)

type ServiceRepository interface {
	Create(service *domain.Service) error
	GetByID(serviceID string) (*domain.Service, error)
	GetAll() ([]*domain.Service, error)
	Update(service *domain.Service) error
	Delete(serviceID string) error

	IncrementWebHookExecutions(serviceID, hookID string) error
}
