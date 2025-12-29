package sqlite

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twttr/kpuppy-backend/internal/domain"
)

func setupContentRepoTest(t *testing.T) (*ContentRepository, func()) {
	tmpFile, err := os.CreateTemp("", "test_content_repo_*.db")
	require.NoError(t, err)
	tmpFile.Close()

	db, err := NewDB(tmpFile.Name())
	require.NoError(t, err)

	err = RunMigrations(db)
	require.NoError(t, err)

	repo := NewContentRepository(db)

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return repo, cleanup
}

func TestContentRepository_GetOrCreate_CreatesNew(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	content, err := repo.GetOrCreate(ctx, 12345)
	require.NoError(t, err)

	assert.NotEmpty(t, content.ID)
	assert.Equal(t, int64(12345), content.KinopubItemID)
}

func TestContentRepository_GetOrCreate_ReturnsExisting(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	content1, err := repo.GetOrCreate(ctx, 12345)
	require.NoError(t, err)

	content2, err := repo.GetOrCreate(ctx, 12345)
	require.NoError(t, err)

	assert.Equal(t, content1.ID, content2.ID)
}

func TestContentRepository_GetByID(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	created, err := repo.GetOrCreate(ctx, 12345)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)

	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.KinopubItemID, found.KinopubItemID)
}

func TestContentRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrContentNotFound, err)
}

func TestContentRepository_GetByKinopubItemID(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	created, err := repo.GetOrCreate(ctx, 12345)
	require.NoError(t, err)

	found, err := repo.GetByKinopubItemID(ctx, 12345)
	require.NoError(t, err)

	assert.Equal(t, created.ID, found.ID)
}

func TestContentRepository_GetByKinopubItemID_NotFound(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	_, err := repo.GetByKinopubItemID(ctx, 99999)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrContentNotFound, err)
}

func TestContentRepository_Create(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	content := &domain.Content{
		ID:            "test-id",
		KinopubItemID: 54321,
	}

	err := repo.Create(ctx, content)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, "test-id")
	require.NoError(t, err)

	assert.Equal(t, int64(54321), found.KinopubItemID)
}

func TestContentRepository_Create_DuplicateID(t *testing.T) {
	repo, cleanup := setupContentRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	content := &domain.Content{
		ID:            "test-id",
		KinopubItemID: 54321,
	}

	err := repo.Create(ctx, content)
	require.NoError(t, err)

	err = repo.Create(ctx, content)
	assert.Error(t, err)
}
