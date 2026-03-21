package service

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/transport/dto"
)

type ServiceManager interface {
	CreateService(response dto.ServiceResponse) (*domain.Service, error)
	GetService(id string) (*domain.Service, error)
	DeleteService(id string) error
	GetAllServices() ([]*domain.Service, error)
	ExecuteWebHook(serviceID, path, method string) (*domain.WebHook, error)
}
