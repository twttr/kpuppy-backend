package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twttr/kpuppy-backend/internal/domain"
)

func setupTestDB(t *testing.T) (*UserRepository, func()) {
	tmpFile, err := os.CreateTemp("", "test_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	db, err := NewDB(tmpFile.Name())
	require.NoError(t, err)

	err = RunMigrations(db)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return NewUserRepository(db), cleanup
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:              "user-123",
		KinopubUsername: "testuser",
		Avatar:          strPtr("https://example.com/avatar.jpg"),
		IsBanned:        false,
		CreatedAt:       time.Now().Truncate(time.Second),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, "user-123")
	require.NoError(t, err)

	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.KinopubUsername, retrieved.KinopubUsername)
	assert.Equal(t, *user.Avatar, *retrieved.Avatar)
	assert.Equal(t, user.IsBanned, retrieved.IsBanned)
}

func TestUserRepository_GetByKinopubUsername(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:              "user-123",
		KinopubUsername: "uniqueuser",
		CreatedAt:       time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetByKinopubUsername(ctx, "uniqueuser")
	require.NoError(t, err)

	assert.Equal(t, user.ID, retrieved.ID)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	assert.Equal(t, domain.ErrUserNotFound, err)
}

func TestUserRepository_UpdateAvatar(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:              "user-123",
		KinopubUsername: "testuser",
		Avatar:          nil,
		CreatedAt:       time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	newAvatar := "https://new-avatar.com/img.jpg"
	err = repo.UpdateAvatar(ctx, "user-123", &newAvatar)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, "user-123")
	require.NoError(t, err)
	assert.Equal(t, newAvatar, *retrieved.Avatar)
}

func TestUserRepository_SetBanned(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:              "user-123",
		KinopubUsername: "testuser",
		IsBanned:        false,
		CreatedAt:       time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	err = repo.SetBanned(ctx, "user-123", true)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, "user-123")
	require.NoError(t, err)
	assert.True(t, retrieved.IsBanned)
}

func TestUserRepository_List(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		user := &domain.User{
			ID:              "user-" + string(rune('a'+i)),
			KinopubUsername: "user" + string(rune('a'+i)),
			CreatedAt:       time.Now().Add(time.Duration(i) * time.Second),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	users, total, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, users, 3)
}

func strPtr(s string) *string {
	return &s
}
