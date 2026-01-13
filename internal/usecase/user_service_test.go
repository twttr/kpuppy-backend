package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twttr/kpuppy-backend/internal/domain"
	"github.com/twttr/kpuppy-backend/internal/repository/mocks"
)

const validHash = "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"

func TestUserService_Provision_NewUser(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	req := &domain.ProvisionRequest{
		UserHash: validHash,
		Avatar:   strPtr("https://example.com/avatar.jpg"),
	}

	userRepo.On("GetByUserHash", ctx, validHash).Return(nil, domain.ErrUserNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := service.Provision(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, validHash, user.UserHash)
	assert.NotEmpty(t, user.DisplayName)
	assert.Equal(t, "https://example.com/avatar.jpg", *user.Avatar)
	assert.False(t, user.IsBanned)
	userRepo.AssertExpectations(t)
}

func TestUserService_Provision_NewUser_DisplayNameGenerated(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	req := &domain.ProvisionRequest{
		UserHash: validHash,
	}

	userRepo.On("GetByUserHash", ctx, validHash).Return(nil, domain.ErrUserNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := service.Provision(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	expectedDisplayName := domain.GeneratePseudonym(validHash)
	assert.Equal(t, expectedDisplayName, user.DisplayName)
	userRepo.AssertExpectations(t)
}

func TestUserService_Provision_ExistingUser(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:          "existing-id",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		Avatar:      nil,
		IsBanned:    false,
		CreatedAt:   time.Now(),
	}

	req := &domain.ProvisionRequest{
		UserHash: validHash,
	}

	userRepo.On("GetByUserHash", ctx, validHash).Return(existingUser, nil)

	user, err := service.Provision(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "existing-id", user.ID)
	assert.Equal(t, "Brave Tiger 42", user.DisplayName)
	userRepo.AssertExpectations(t)
}

func TestUserService_Provision_ExistingUserWithAvatarUpdate(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:          "existing-id",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		Avatar:      nil,
		IsBanned:    false,
		CreatedAt:   time.Now(),
	}

	newAvatar := "https://example.com/new-avatar.jpg"
	req := &domain.ProvisionRequest{
		UserHash: validHash,
		Avatar:   &newAvatar,
	}

	userRepo.On("GetByUserHash", ctx, validHash).Return(existingUser, nil)
	userRepo.On("UpdateAvatar", ctx, "existing-id", &newAvatar).Return(nil)

	user, err := service.Provision(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, &newAvatar, user.Avatar)
	userRepo.AssertExpectations(t)
}

func TestUserService_Provision_EmptyHash(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	req := &domain.ProvisionRequest{
		UserHash: "",
	}

	user, err := service.Provision(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, domain.ErrUserHashEmpty, err)
}

func TestUserService_Provision_InvalidHash(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	req := &domain.ProvisionRequest{
		UserHash: "tooshort",
	}

	user, err := service.Provision(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, domain.ErrInvalidUserHash, err)
}

func TestUserService_SetBanned_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:          "user-id",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		IsBanned:    false,
	}

	userRepo.On("GetByID", ctx, "user-id").Return(existingUser, nil)
	userRepo.On("SetBanned", ctx, "user-id", true).Return(nil)

	err := service.SetBanned(ctx, "user-id", true)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_SetBanned_UserNotFound(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrUserNotFound)

	err := service.SetBanned(ctx, "nonexistent", true)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_List(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	users := []domain.User{
		{ID: "1", UserHash: validHash, DisplayName: "Brave Tiger 42"},
		{ID: "2", UserHash: "b8c4d3e2f9e0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b6", DisplayName: "Calm Owl 15"},
	}

	userRepo.On("List", ctx, 0, 20).Return(users, 2, nil)

	result, total, err := service.List(ctx, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
	userRepo.AssertExpectations(t)
}

func TestUserService_GetByID(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:          "user-id",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(existingUser, nil)

	user, err := service.GetByID(ctx, "user-id")

	assert.NoError(t, err)
	assert.Equal(t, "Brave Tiger 42", user.DisplayName)
	userRepo.AssertExpectations(t)
}

func strPtr(s string) *string {
	return &s
}
