package service

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/service/jwt"
	"ServiceManager/internal/transport/dto"
	"context"
)

type ServiceManager interface {
	CreateService(ctx context.Context, response dto.ServiceResponse) (*domain.Service, error)
	GetService(ctx context.Context, id string) (*domain.Service, error)
	UpdateService(ctx context.Context, response dto.ServiceResponse) (*domain.Service, error)
	DeleteService(ctx context.Context, id string) error
	GetAllServices(ctx context.Context) ([]*domain.Service, error)

	IncrementWebHook(ctx context.Context, serviceID, webhookID string) bool
}

type ServiceEmail interface {
	SendOTP(ctx context.Context, email, code, purpose string) error
}

type ServiceJWT interface {
	GenerateAccessToken(userID, email string) (string, error)
	GenerateRefreshToken(userID, email string) (string, error)
	ParseToken(tokenString string) (*jwt.Claims, error)
}
