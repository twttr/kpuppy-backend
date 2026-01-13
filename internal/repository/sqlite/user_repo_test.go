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

const validHash = "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5"

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
		ID:          "user-123",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		Avatar:      strPtr("https://example.com/avatar.jpg"),
		IsBanned:    false,
		CreatedAt:   time.Now().Truncate(time.Second),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, "user-123")
	require.NoError(t, err)

	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.UserHash, retrieved.UserHash)
	assert.Equal(t, user.DisplayName, retrieved.DisplayName)
	assert.Equal(t, *user.Avatar, *retrieved.Avatar)
	assert.Equal(t, user.IsBanned, retrieved.IsBanned)
}

func TestUserRepository_GetByUserHash(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	retrieved, err := repo.GetByUserHash(ctx, validHash)
	require.NoError(t, err)

	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.DisplayName, retrieved.DisplayName)
}

func TestUserRepository_GetByUserHash_NotFound(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := repo.GetByUserHash(ctx, "nonexistent0000000000000000000000000000000000000000000000000000")
	assert.Equal(t, domain.ErrUserNotFound, err)
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
		ID:          "user-123",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		Avatar:      nil,
		CreatedAt:   time.Now(),
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
		ID:          "user-123",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		IsBanned:    false,
		CreatedAt:   time.Now(),
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

	hashes := []string{
		"a000000000000000000000000000000000000000000000000000000000000000",
		"b000000000000000000000000000000000000000000000000000000000000000",
		"c000000000000000000000000000000000000000000000000000000000000000",
		"d000000000000000000000000000000000000000000000000000000000000000",
		"e000000000000000000000000000000000000000000000000000000000000000",
	}

	for i := 0; i < 5; i++ {
		user := &domain.User{
			ID:          "user-" + string(rune('a'+i)),
			UserHash:    hashes[i],
			DisplayName: "User " + string(rune('A'+i)),
			CreatedAt:   time.Now().Add(time.Duration(i) * time.Second),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	users, total, err := repo.List(ctx, 0, 3)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, users, 3)
}

func TestUserRepository_List_IncludesDisplayName(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    validHash,
		DisplayName: "Brave Tiger 42",
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	users, _, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Brave Tiger 42", users[0].DisplayName)
}

func strPtr(s string) *string {
	return &s
}
