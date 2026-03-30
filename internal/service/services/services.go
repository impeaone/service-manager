package services

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/repository"
	"ServiceManager/internal/transport/dto"
	"ServiceManager/pkg/utils"
	"context"
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
	if err := s.repo.Create(context.TODO(), service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *ServiceManager) UpdateService(resp dto.ServiceResponse) (*domain.Service, error) {
	webhooks := make([]domain.WebHook, len(resp.WebHooks))

	for i, respHook := range resp.WebHooks {
		var id = respHook.ID
		// TODO норм валидацию надо
		if respHook.ID == "" || respHook.ID == "null" {
			id = utils.GenerateUUID()
		}

		webhooks[i] = domain.WebHook{
			ID:         id,
			Name:       respHook.Name,
			Path:       respHook.Path,
			Type:       domain.WebHookType(respHook.Type),
			Method:     respHook.Method,
			Executions: respHook.Executions,
			LastCall:   time.Now(),
		}
	}
	service := &domain.Service{
		ID:        resp.ID,
		Name:      resp.Name,
		Status:    domain.ServiceStatus(resp.Status),
		WebHooks:  webhooks,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Update(context.TODO(), service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *ServiceManager) GetService(id string) (*domain.Service, error) {
	return s.repo.GetByID(context.TODO(), id)
}

func (s *ServiceManager) GetAllServices() ([]*domain.Service, error) {
	return s.repo.GetAll(context.TODO())
}

func (s *ServiceManager) DeleteService(serviceID string) error {
	return s.repo.Delete(context.TODO(), serviceID)
}

func (s *ServiceManager) IncrementWebHook(serviceID, webhookID string) bool {
	service, err := s.repo.GetByID(context.TODO(), serviceID)
	if err != nil {
		return false
	}

	var targetHook *domain.WebHook
	for _, hook := range service.WebHooks {
		if hook.ID == webhookID {
			targetHook = &hook
			break
		}
	}

	if targetHook == nil {
		return false
	}

	if err = s.repo.IncrementWebHookExecutions(context.TODO(), serviceID, targetHook.ID); err != nil {
		return false
	}

	return true
}
