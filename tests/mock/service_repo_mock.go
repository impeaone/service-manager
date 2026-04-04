package mocks

import (
	"ServiceManager/internal/domain"
	"context"
	"github.com/stretchr/testify/mock"
)

type MockServiceRepository struct {
	mock.Mock
}

func (m *MockServiceRepository) Create(ctx context.Context, userID string, service *domain.Service) error {
	args := m.Called(ctx, userID, service)
	return args.Error(0)
}

func (m *MockServiceRepository) GetByID(ctx context.Context, userID, serviceID string) (*domain.Service, error) {
	args := m.Called(ctx, userID, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Service), args.Error(1)
}

func (m *MockServiceRepository) GetAll(ctx context.Context, userID string) ([]*domain.Service, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Service), args.Error(1)
}

func (m *MockServiceRepository) Update(ctx context.Context, userID string, service *domain.Service) error {
	args := m.Called(ctx, userID, service)
	return args.Error(0)
}

func (m *MockServiceRepository) Delete(ctx context.Context, userID string, serviceID string) error {
	args := m.Called(ctx, userID, serviceID)
	return args.Error(0)
}

func (m *MockServiceRepository) IncrementWebHookExecutions(ctx context.Context, serviceID, hookID string) error {
	args := m.Called(ctx, serviceID, hookID)
	return args.Error(0)
}
