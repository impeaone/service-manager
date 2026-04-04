package services

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/middleware"
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

func (s *ServiceManager) CreateService(ctx context.Context, resp dto.ServiceResponse) (*domain.Service, error) {
	webhooks := make([]domain.WebHook, len(resp.WebHooks))

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

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
	if err := s.repo.Create(ctx, user.ID, service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *ServiceManager) UpdateService(ctx context.Context, resp dto.ServiceResponse) (*domain.Service, error) {
	webhooks := make([]domain.WebHook, len(resp.WebHooks))

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

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

	if err := s.repo.Update(ctx, user.ID, service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *ServiceManager) GetService(ctx context.Context, id string) (*domain.Service, error) {

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, user.ID, id)
}

func (s *ServiceManager) GetAllServices(ctx context.Context) ([]*domain.Service, error) {

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.GetAll(ctx, user.ID)
}

func (s *ServiceManager) DeleteService(ctx context.Context, serviceID string) error {

	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, user.ID, serviceID)
}

func (s *ServiceManager) IncrementWebHook(ctx context.Context, serviceID, webhookID string) bool {
	user, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return false
	}

	service, err := s.repo.GetByID(ctx, user.ID, serviceID)
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

	if err = s.repo.IncrementWebHookExecutions(ctx, serviceID, targetHook.ID); err != nil {
		return false
	}

	return true
}
