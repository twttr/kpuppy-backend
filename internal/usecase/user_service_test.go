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

func TestUserService_Provision_NewUser(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	req := &domain.ProvisionRequest{
		Username: "testuser",
		Avatar:   strPtr("https://example.com/avatar.jpg"),
	}

	userRepo.On("GetByKinopubUsername", ctx, "testuser").Return(nil, domain.ErrUserNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := service.Provision(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.KinopubUsername)
	assert.Equal(t, "https://example.com/avatar.jpg", *user.Avatar)
	assert.False(t, user.IsBanned)
	userRepo.AssertExpectations(t)
}

func TestUserService_Provision_ExistingUser(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:              "existing-id",
		KinopubUsername: "testuser",
		Avatar:          nil,
		IsBanned:        false,
		CreatedAt:       time.Now(),
	}

	req := &domain.ProvisionRequest{
		Username: "testuser",
	}

	userRepo.On("GetByKinopubUsername", ctx, "testuser").Return(existingUser, nil)

	user, err := service.Provision(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "existing-id", user.ID)
	userRepo.AssertExpectations(t)
}

func TestUserService_Provision_ExistingUserWithAvatarUpdate(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:              "existing-id",
		KinopubUsername: "testuser",
		Avatar:          nil,
		IsBanned:        false,
		CreatedAt:       time.Now(),
	}

	newAvatar := "https://example.com/new-avatar.jpg"
	req := &domain.ProvisionRequest{
		Username: "testuser",
		Avatar:   &newAvatar,
	}

	userRepo.On("GetByKinopubUsername", ctx, "testuser").Return(existingUser, nil)
	userRepo.On("UpdateAvatar", ctx, "existing-id", &newAvatar).Return(nil)

	user, err := service.Provision(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, &newAvatar, user.Avatar)
	userRepo.AssertExpectations(t)
}

func TestUserService_Provision_EmptyUsername(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	req := &domain.ProvisionRequest{
		Username: "",
	}

	user, err := service.Provision(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, domain.ErrUsernameEmpty, err)
}

func TestUserService_SetBanned_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	service := NewUserService(userRepo)
	ctx := context.Background()

	existingUser := &domain.User{
		ID:              "user-id",
		KinopubUsername: "testuser",
		IsBanned:        false,
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
		{ID: "1", KinopubUsername: "user1"},
		{ID: "2", KinopubUsername: "user2"},
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
		ID:              "user-id",
		KinopubUsername: "testuser",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(existingUser, nil)

	user, err := service.GetByID(ctx, "user-id")

	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.KinopubUsername)
	userRepo.AssertExpectations(t)
}

func strPtr(s string) *string {
	return &s
}
