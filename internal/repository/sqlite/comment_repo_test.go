package sqlite

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twttr/kpuppy-backend/internal/domain"
)

func setupCommentTestDB(t *testing.T) (*sql.DB, *CommentRepository, *ContentRepository, *UserRepository, func()) {
	tmpFile, err := os.CreateTemp("", "test_comment_*.db")
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

	return db, NewCommentRepository(db), NewContentRepository(db), NewUserRepository(db), cleanup
}

func TestCommentRepository_CreateAndGet(t *testing.T) {
	_, commentRepo, contentRepo, userRepo, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "testuser",
		CreatedAt:   time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	content := &domain.Content{
		ID:            "content-456",
		KinopubItemID: 12345,
		CreatedAt:     time.Now(),
	}
	err = contentRepo.Create(ctx, content)
	require.NoError(t, err)

	comment := &domain.Comment{
		ID:        "comment-789",
		ContentID: "content-456",
		UserID:    "user-123",
		Text:      "Test comment",
		Spoiler:   true,
		CreatedAt: time.Now().Truncate(time.Second),
	}

	err = commentRepo.Create(ctx, comment)
	require.NoError(t, err)

	retrieved, err := commentRepo.GetByID(ctx, "comment-789")
	require.NoError(t, err)

	assert.Equal(t, comment.ID, retrieved.ID)
	assert.Equal(t, comment.ContentID, retrieved.ContentID)
	assert.Equal(t, comment.UserID, retrieved.UserID)
	assert.Equal(t, comment.Text, retrieved.Text)
	assert.True(t, retrieved.Spoiler)
	assert.NotNil(t, retrieved.User)
	assert.Equal(t, "testuser", retrieved.User.DisplayName)
}

func TestCommentRepository_GetByContentID(t *testing.T) {
	_, commentRepo, contentRepo, userRepo, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "testuser",
		CreatedAt:   time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	content := &domain.Content{
		ID:            "content-456",
		KinopubItemID: 12345,
		CreatedAt:     time.Now(),
	}
	err = contentRepo.Create(ctx, content)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		comment := &domain.Comment{
			ID:        "comment-" + string(rune('a'+i)),
			ContentID: "content-456",
			UserID:    "user-123",
			Text:      "Comment " + string(rune('a'+i)),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		err = commentRepo.Create(ctx, comment)
		require.NoError(t, err)
	}

	comments, err := commentRepo.GetByContentID(ctx, "content-456")
	require.NoError(t, err)

	assert.Len(t, comments, 5)
}

func TestCommentRepository_Replies(t *testing.T) {
	_, commentRepo, contentRepo, userRepo, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "testuser",
		CreatedAt:   time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	content := &domain.Content{
		ID:            "content-456",
		KinopubItemID: 12345,
		CreatedAt:     time.Now(),
	}
	err = contentRepo.Create(ctx, content)
	require.NoError(t, err)

	parent := &domain.Comment{
		ID:        "parent-comment",
		ContentID: "content-456",
		UserID:    "user-123",
		Text:      "Parent comment",
		CreatedAt: time.Now(),
	}
	err = commentRepo.Create(ctx, parent)
	require.NoError(t, err)

	parentID := "parent-comment"
	reply := &domain.Comment{
		ID:        "reply-comment",
		ContentID: "content-456",
		UserID:    "user-123",
		Text:      "Reply comment",
		ParentID:  &parentID,
		CreatedAt: time.Now(),
	}
	err = commentRepo.Create(ctx, reply)
	require.NoError(t, err)

	replies, err := commentRepo.GetReplies(ctx, "parent-comment")
	require.NoError(t, err)

	assert.Len(t, replies, 1)
	assert.Equal(t, "Reply comment", replies[0].Text)
}

func TestCommentRepository_SoftDelete(t *testing.T) {
	_, commentRepo, contentRepo, userRepo, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "testuser",
		CreatedAt:   time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	content := &domain.Content{
		ID:            "content-456",
		KinopubItemID: 12345,
		CreatedAt:     time.Now(),
	}
	err = contentRepo.Create(ctx, content)
	require.NoError(t, err)

	comment := &domain.Comment{
		ID:        "comment-789",
		ContentID: "content-456",
		UserID:    "user-123",
		Text:      "Test comment",
		CreatedAt: time.Now(),
	}
	err = commentRepo.Create(ctx, comment)
	require.NoError(t, err)

	err = commentRepo.SoftDelete(ctx, "comment-789")
	require.NoError(t, err)

	retrieved, err := commentRepo.GetByID(ctx, "comment-789")
	require.NoError(t, err)
	assert.True(t, retrieved.IsDeleted())
}

func TestCommentRepository_Update(t *testing.T) {
	_, commentRepo, contentRepo, userRepo, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "testuser",
		CreatedAt:   time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	content := &domain.Content{
		ID:            "content-456",
		KinopubItemID: 12345,
		CreatedAt:     time.Now(),
	}
	err = contentRepo.Create(ctx, content)
	require.NoError(t, err)

	comment := &domain.Comment{
		ID:        "comment-789",
		ContentID: "content-456",
		UserID:    "user-123",
		Text:      "Original text",
		Spoiler:   false,
		CreatedAt: time.Now(),
	}
	err = commentRepo.Create(ctx, comment)
	require.NoError(t, err)

	now := time.Now()
	comment.Text = "Updated text"
	comment.Spoiler = true
	comment.EditedAt = &now

	err = commentRepo.Update(ctx, comment)
	require.NoError(t, err)

	retrieved, err := commentRepo.GetByID(ctx, "comment-789")
	require.NoError(t, err)

	assert.Equal(t, "Updated text", retrieved.Text)
	assert.True(t, retrieved.Spoiler)
	assert.NotNil(t, retrieved.EditedAt)
}

func TestCommentRepository_List(t *testing.T) {
	_, commentRepo, contentRepo, userRepo, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "testuser",
		CreatedAt:   time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	content := &domain.Content{
		ID:            "content-456",
		KinopubItemID: 12345,
		CreatedAt:     time.Now(),
	}
	err = contentRepo.Create(ctx, content)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		comment := &domain.Comment{
			ID:        "comment-" + string(rune('a'+i)),
			ContentID: "content-456",
			UserID:    "user-123",
			Text:      "Comment " + string(rune('a'+i)),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		err = commentRepo.Create(ctx, comment)
		require.NoError(t, err)
	}

	comments, total, err := commentRepo.List(ctx, 0, 3, false)
	require.NoError(t, err)

	assert.Equal(t, 5, total)
	assert.Len(t, comments, 3)
}

func TestCommentRepository_List_IncludeDeleted(t *testing.T) {
	_, commentRepo, contentRepo, userRepo, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	user := &domain.User{
		ID:          "user-123",
		UserHash:    "a7b3c2f1e8d9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5",
		DisplayName: "testuser",
		CreatedAt:   time.Now(),
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	content := &domain.Content{
		ID:            "content-456",
		KinopubItemID: 12345,
		CreatedAt:     time.Now(),
	}
	err = contentRepo.Create(ctx, content)
	require.NoError(t, err)

	comment1 := &domain.Comment{
		ID:        "comment-1",
		ContentID: "content-456",
		UserID:    "user-123",
		Text:      "Active comment",
		CreatedAt: time.Now(),
	}
	err = commentRepo.Create(ctx, comment1)
	require.NoError(t, err)

	comment2 := &domain.Comment{
		ID:        "comment-2",
		ContentID: "content-456",
		UserID:    "user-123",
		Text:      "Deleted comment",
		CreatedAt: time.Now(),
	}
	err = commentRepo.Create(ctx, comment2)
	require.NoError(t, err)

	err = commentRepo.SoftDelete(ctx, "comment-2")
	require.NoError(t, err)

	commentsWithoutDeleted, total1, err := commentRepo.List(ctx, 0, 10, false)
	require.NoError(t, err)
	assert.Equal(t, 1, total1)
	assert.Len(t, commentsWithoutDeleted, 1)

	commentsWithDeleted, total2, err := commentRepo.List(ctx, 0, 10, true)
	require.NoError(t, err)
	assert.Equal(t, 2, total2)
	assert.Len(t, commentsWithDeleted, 2)
}

func TestCommentRepository_GetByID_NotFound(t *testing.T) {
	_, commentRepo, _, _, cleanup := setupCommentTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := commentRepo.GetByID(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentNotFound, err)
}
