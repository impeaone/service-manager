// tests/webhook_service_test.go
package tests

import (
	"ServiceManager/internal/middleware"
	"ServiceManager/internal/service/services"
	mocks "ServiceManager/tests/mock"
	"context"
	"errors"
	"testing"

	"ServiceManager/internal/domain"
	"ServiceManager/internal/transport/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceManager_CreateService(t *testing.T) {
	mockServiceRepo := new(mocks.MockServiceRepository)

	serviceManager := services.NewServiceManager(mockServiceRepo)

	user := &domain.User{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	ctx := context.WithValue(context.Background(), middleware.UserContextKey, user)

	tests := []struct {
		name        string
		request     dto.ServiceResponse
		setupMock   func()
		expectedErr error
	}{
		{
			name: "успешное создание сервиса",
			request: dto.ServiceResponse{
				Name:   "Мой сервис",
				Status: "active",
			},
			setupMock: func() {
				mockServiceRepo.On("Create", mock.Anything, user.ID, mock.AnythingOfType("*domain.Service")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "ошибка при создании",
			request: dto.ServiceResponse{
				Name: "Тест",
			},
			setupMock: func() {
				mockServiceRepo.On("Create", mock.Anything, user.ID, mock.AnythingOfType("*domain.Service")).Return(errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServiceRepo.ExpectedCalls = nil

			tt.setupMock()

			result, err := serviceManager.CreateService(ctx, tt.request)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Name, result.Name)
			}

			mockServiceRepo.AssertExpectations(t)
		})
	}
}
