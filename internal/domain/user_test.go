package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvisionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     ProvisionRequest
		wantErr error
	}{
		{
			name:    "valid request",
			req:     ProvisionRequest{Username: "testuser"},
			wantErr: nil,
		},
		{
			name:    "valid request with avatar",
			req:     ProvisionRequest{Username: "testuser", Avatar: strPtr("https://example.com/avatar.jpg")},
			wantErr: nil,
		},
		{
			name:    "empty username",
			req:     ProvisionRequest{Username: ""},
			wantErr: ErrUsernameEmpty,
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
		ID:              "user-123",
		KinopubUsername: "testuser",
		Avatar:          &avatar,
	}

	resp := user.ToResponse()

	assert.Equal(t, "user-123", resp.ID)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, &avatar, resp.Avatar)
}

func TestUser_ToResponse_NoAvatar(t *testing.T) {
	user := &User{
		ID:              "user-123",
		KinopubUsername: "testuser",
		Avatar:          nil,
	}

	resp := user.ToResponse()

	assert.Equal(t, "user-123", resp.ID)
	assert.Equal(t, "testuser", resp.Username)
	assert.Nil(t, resp.Avatar)
}
