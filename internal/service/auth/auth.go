package service

import (
	"ServiceManager/internal/domain"
	"ServiceManager/internal/repository"
	"ServiceManager/internal/service/jwt"

	"ServiceManager/internal/service"

	"context"
	"fmt"
	"time"

	redisRepo "ServiceManager/internal/repository/redis"
)

type AuthService struct {
	userRepo     repository.UserRepository
	otpRepo      *redisRepo.OTPRepository
	tokenRepo    *redisRepo.TokenRepository
	jwtService   *jwt.JWTService
	emailService service.ServiceEmail
}

func NewAuthService(
	userRepo repository.UserRepository,
	otpRepo *redisRepo.OTPRepository,
	tokenRepo *redisRepo.TokenRepository,
	jwtService *jwt.JWTService,
	emailService service.ServiceEmail,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		otpRepo:      otpRepo,
		tokenRepo:    tokenRepo,
		jwtService:   jwtService,
		emailService: emailService,
	}
}

func (s *AuthService) RequestRegistrationOTP(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser != nil {
		return fmt.Errorf("user already exists")
	}

	code, err := s.otpRepo.GenerateOTP(ctx, email, "registration", 10*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err = s.emailService.SendOTP(ctx, email, code, "registration"); err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	return nil
}

func (s *AuthService) VerifyRegistrationOTP(ctx context.Context, email, code, name string) (*domain.AuthResponse, error) {
	valid, err := s.otpRepo.VerifyOTP(ctx, email, "registration", code)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	user := &domain.User{
		Email:      email,
		Name:       name,
		IsVerified: true,
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err = s.tokenRepo.SaveRefreshToken(ctx, user.ID, refreshToken, s.jwtService.RefreshTTL); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &domain.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RequestLoginOTP(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	code, err := s.otpRepo.GenerateOTP(ctx, email, "login", 10*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err = s.emailService.SendOTP(ctx, email, code, "login"); err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	return nil
}

func (s *AuthService) VerifyLoginOTP(ctx context.Context, email, code string) (*domain.AuthResponse, error) {
	valid, err := s.otpRepo.VerifyOTP(ctx, email, "login", code)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err = s.tokenRepo.SaveRefreshToken(ctx, user.ID, refreshToken, s.jwtService.RefreshTTL); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &domain.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	userID, err := s.tokenRepo.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)

	newAccessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err = s.tokenRepo.SaveRefreshToken(ctx, user.ID, newRefreshToken, s.jwtService.RefreshTTL); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &domain.AuthResponse{
		User:         *user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *AuthService) VerifyToken(ctx context.Context, tokenString string) (*domain.User, error) {
	claims, err := s.jwtService.ParseToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
