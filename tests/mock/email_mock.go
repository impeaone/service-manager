package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockServiceEmail struct {
	mock.Mock
}

func (m *MockServiceEmail) SendOTP(ctx context.Context, email, code, purpose string) error {
	args := m.Called(ctx, email, code, purpose)
	return args.Error(0)
}
