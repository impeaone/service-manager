package service_manager

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/repository"
	"ServiceManager/internal/transport/dto"
	"ServiceManager/pkg/utils"
	"time"
)

type ServiceManager struct {
	repo repository.ServiceRepository
}

func NewServiceManager(repo repository.ServiceRepository) *ServiceManager {
	return &ServiceManager{repo}
}

func (s *ServiceManager) CreateService(resp dto.ServiceResponse) (*domain.Service, error) {
	webhooks := make([]domain.WebHook, len(resp.WebHooks))
	for i, respHook := range resp.WebHooks {
		webhooks[i] = domain.WebHook{
			ID:         utils.GenerateUUID(),
			Name:       respHook.Name,
			Path:       respHook.Path,
			Type:       domain.WebHookType(respHook.Type),
			Method:     respHook.Method,
			Executions: 0,
			LastCall:   time.Now(),
		}
	}

	service := &domain.Service{
		ID:        utils.GenerateUUID(),
		Name:      resp.Name,
		Status:    domain.ServiceStatus(resp.Status),
		WebHooks:  webhooks,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *ServiceManager) GetService(id string) (*domain.Service, error) {
	return s.repo.GetByID(id)
}

func (s *ServiceManager) GetAllServices() ([]*domain.Service, error) {
	return s.repo.GetAll()
}

func (s *ServiceManager) DeleteService(serviceID string) error {
	return s.repo.Delete(serviceID)
}

func (s *ServiceManager) ExecuteWebHook(serviceID, path, method string) (*domain.WebHook, error) {
	service, err := s.repo.GetByID(serviceID)
	if err != nil {
		return nil, err
	}

	// Находим веб-хук по пути
	var targetHook *domain.WebHook
	for _, hook := range service.WebHooks {
		if hook.Path == path && hook.Method == method {
			targetHook = &hook
			break
		}
	}

	if targetHook == nil {
		return nil, repository.ServiceNotFoundError
	}

	// Инкрементируем счетчик вызовов
	if err = s.repo.IncrementWebHookExecutions(serviceID, targetHook.ID); err != nil {
		return nil, err
	}

	return targetHook, nil
}
