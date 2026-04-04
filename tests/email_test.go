package tests

import (
	mocks "ServiceManager/tests/mock"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendOTP_Success(t *testing.T) {
	mockEmail := new(mocks.MockServiceEmail)
	ctx := context.Background()

	mockEmail.On("SendOTP", ctx, "user@example.com", "123456", "login").Return(nil)

	err := mockEmail.SendOTP(ctx, "user@example.com", "123456", "login")

	assert.NoError(t, err)
	mockEmail.AssertExpectations(t)
}

func TestSendOTP_Error(t *testing.T) {
	mockEmail := new(mocks.MockServiceEmail)
	ctx := context.Background()

	mockEmail.On("SendOTP", ctx, "invalid@example.com", "123456", "login").Return(errors.New("smtp error"))

	err := mockEmail.SendOTP(ctx, "invalid@example.com", "123456", "login")

	assert.Error(t, err)
	mockEmail.AssertExpectations(t)
}
