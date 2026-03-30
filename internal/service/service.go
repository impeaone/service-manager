package service

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/service/jwt"
	"ServiceManager/internal/transport/dto"
	"context"
)

type ServiceManager interface {
	CreateService(response dto.ServiceResponse) (*domain.Service, error)
	GetService(id string) (*domain.Service, error)
	UpdateService(response dto.ServiceResponse) (*domain.Service, error)
	DeleteService(id string) error
	GetAllServices() ([]*domain.Service, error)

	IncrementWebHook(serviceID, webhookID string) bool
}

type ServiceEmail interface {
	SendOTP(ctx context.Context, email, code, purpose string) error
}

type ServiceJWT interface {
	GenerateAccessToken(userID, email string) (string, error)
	GenerateRefreshToken(userID, email string) (string, error)
	ParseToken(tokenString string) (*jwt.Claims, error)
}
