package email

import (
	"context"
	"fmt"
)

type MockEmailService struct{}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{}
}

func (s *MockEmailService) SendOTP(_ context.Context, email, code, purpose string) error {
	fmt.Printf("Sending OTP %s to %s for %s\n", code, email, purpose)
	return nil
}
