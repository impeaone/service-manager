package mocks

import (
	"context"

	"ServiceManager/internal/domain"
	"ServiceManager/internal/transport/dto"
	"github.com/stretchr/testify/mock"
)

type MockServiceManager struct {
	mock.Mock
}

func (m *MockServiceManager) CreateService(ctx context.Context, response dto.ServiceResponse) (*domain.Service, error) {
	args := m.Called(ctx, response)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Service), args.Error(1)
}

func (m *MockServiceManager) GetService(ctx context.Context, id string) (*domain.Service, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Service), args.Error(1)
}

func (m *MockServiceManager) UpdateService(ctx context.Context, response dto.ServiceResponse) (*domain.Service, error) {
	args := m.Called(ctx, response)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Service), args.Error(1)
}

func (m *MockServiceManager) DeleteService(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceManager) GetAllServices(ctx context.Context) ([]*domain.Service, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Service), args.Error(1)
}

func (m *MockServiceManager) IncrementWebHook(ctx context.Context, serviceID, webhookID string) bool {
	args := m.Called(ctx, serviceID, webhookID)
	return args.Bool(0)
}
