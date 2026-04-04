// tests/jwt_test.go
package tests

import (
	"testing"

	"ServiceManager/tests/mock"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestParseToken_Success(t *testing.T) {
	mockJWT := new(mocks.MockServiceJWT)

	claims := jwt.MapClaims{
		"user_id": "123",
		"email":   "test@example.com",
	}

	var claimsInterface jwt.Claims = claims

	mockJWT.On("ParseToken", "valid.token").Return(&claimsInterface, nil)

	result, err := mockJWT.ParseToken("valid.token")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	mockJWT.AssertExpectations(t)
}

func TestParseToken_Error(t *testing.T) {
	mockJWT := new(mocks.MockServiceJWT)

	mockJWT.On("ParseToken", "invalid.token").Return(nil, jwt.ErrTokenMalformed)

	result, err := mockJWT.ParseToken("invalid.token")

	assert.Error(t, err)
	assert.Nil(t, result)
	mockJWT.AssertExpectations(t)
}
