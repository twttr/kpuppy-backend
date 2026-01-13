package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIError(t *testing.T) {
	err := errors.New("test error")
	apiErr := NewAPIError(err, CodeValidationError)

	assert.Equal(t, "test error", apiErr.Error)
	assert.Equal(t, CodeValidationError, apiErr.Code)
}

func TestNewAPIError_WithDomainError(t *testing.T) {
	apiErr := NewAPIError(ErrUserNotFound, CodeUserNotProvisioned)

	assert.Equal(t, "user not found", apiErr.Error)
	assert.Equal(t, CodeUserNotProvisioned, apiErr.Code)
}

func TestErrorCodes(t *testing.T) {
	assert.Equal(t, ErrorCode("USER_NOT_PROVISIONED"), CodeUserNotProvisioned)
	assert.Equal(t, ErrorCode("CONTENT_NOT_FOUND"), CodeContentNotFound)
	assert.Equal(t, ErrorCode("COMMENT_NOT_FOUND"), CodeCommentNotFound)
	assert.Equal(t, ErrorCode("VALIDATION_ERROR"), CodeValidationError)
	assert.Equal(t, ErrorCode("FORBIDDEN"), CodeForbidden)
	assert.Equal(t, ErrorCode("USER_BANNED"), CodeUserBanned)
	assert.Equal(t, ErrorCode("RATE_LIMIT_EXCEEDED"), CodeRateLimitExceeded)
}

func TestDomainErrors(t *testing.T) {
	assert.NotNil(t, ErrUserNotFound)
	assert.NotNil(t, ErrUserBanned)
	assert.NotNil(t, ErrUserHashEmpty)
	assert.NotNil(t, ErrInvalidUserHash)
	assert.NotNil(t, ErrContentNotFound)
	assert.NotNil(t, ErrCommentNotFound)
	assert.NotNil(t, ErrCommentDeleted)
	assert.NotNil(t, ErrNotCommentOwner)
	assert.NotNil(t, ErrCommentTooLong)
	assert.NotNil(t, ErrCommentEmpty)
	assert.NotNil(t, ErrUserNotProvisioned)
	assert.NotNil(t, ErrRateLimitExceeded)
}
