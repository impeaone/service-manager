package memory_repository

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/repository"
	"context"
	"sync"
	"time"
)

type MemoryRepository struct {
	services map[string]*domain.Service
	mu       sync.RWMutex
}

func NewMemoryRepository() (*MemoryRepository, error) {
	return &MemoryRepository{
		services: make(map[string]*domain.Service),
	}, nil
}

func (r *MemoryRepository) Create(service *domain.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services[service.ID] = service
	return nil
}

func (r *MemoryRepository) GetByID(serviceID string) (*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, ok := r.services[serviceID]
	if !ok {
		return nil, repository.ServiceNotFoundError
	}
	return service, nil
}

func (r *MemoryRepository) GetAll() ([]*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*domain.Service, 0, len(r.services))

	for _, service := range r.services {
		services = append(services, service)
	}
	return services, nil
}

func (r *MemoryRepository) Update(service *domain.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.services[service.ID]; !ok {
		return repository.ServiceNotFoundError
	}

	service.UpdatedAt = time.Now()
	r.services[service.ID] = service
	return nil
}

func (r *MemoryRepository) Delete(serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.services[serviceID]; !ok {
		return repository.ServiceNotFoundError
	}
	delete(r.services, serviceID)
	return nil
}

func (r *MemoryRepository) IncrementWebHookExecutions(serviceID, hookID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	service, exists := r.services[serviceID]
	if !exists {
		return repository.ServiceNotFoundError
	}

	for i, hook := range service.WebHooks {
		if hook.ID == hookID {
			service.WebHooks[i].Executions++
			service.WebHooks[i].LastCall = time.Now()
			service.UpdatedAt = time.Now()
			return nil
		}
	}

	return repository.ServiceNotFoundError
}

func (r *MemoryRepository) Stop(_ context.Context) error {
	return nil
}
