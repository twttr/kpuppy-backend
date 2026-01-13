package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvisionRequest_Validate(t *testing.T) {
	validHash := "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"

	tests := []struct {
		name    string
		req     ProvisionRequest
		wantErr error
	}{
		{
			name:    "valid request",
			req:     ProvisionRequest{UserHash: validHash},
			wantErr: nil,
		},
		{
			name:    "valid request with avatar",
			req:     ProvisionRequest{UserHash: validHash, Avatar: strPtr("https://example.com/avatar.jpg")},
			wantErr: nil,
		},
		{
			name:    "empty hash",
			req:     ProvisionRequest{UserHash: ""},
			wantErr: ErrUserHashEmpty,
		},
		{
			name:    "invalid hash length - too short",
			req:     ProvisionRequest{UserHash: "abc123"},
			wantErr: ErrInvalidUserHash,
		},
		{
			name:    "invalid hash length - too long",
			req:     ProvisionRequest{UserHash: validHash + "extra"},
			wantErr: ErrInvalidUserHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestUser_ToResponse(t *testing.T) {
	avatar := "https://example.com/avatar.jpg"
	user := &User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "Brave Tiger 42",
		Avatar:      &avatar,
	}

	resp := user.ToResponse()

	assert.Equal(t, "user-123", resp.ID)
	assert.Equal(t, "Brave Tiger 42", resp.DisplayName)
	assert.Equal(t, &avatar, resp.Avatar)
}

func TestUser_ToResponse_NoAvatar(t *testing.T) {
	user := &User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "Calm Owl 15",
		Avatar:      nil,
	}

	resp := user.ToResponse()

	assert.Equal(t, "user-123", resp.ID)
	assert.Equal(t, "Calm Owl 15", resp.DisplayName)
	assert.Nil(t, resp.Avatar)
}