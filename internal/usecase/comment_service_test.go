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

func TestCommentService_GetComments_Success(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	content := &domain.Content{
		ID:            "content-id",
		KinopubItemID: 12345,
	}

	comments := []domain.Comment{
		{ID: "c1", Text: "comment 1", ContentID: "content-id"},
		{ID: "c2", Text: "comment 2", ContentID: "content-id"},
	}

	contentRepo.On("GetByKinopubItemID", ctx, int64(12345)).Return(content, nil)
	commentRepo.On("GetByContentID", ctx, "content-id").Return(comments, nil)

	result, err := service.GetComments(ctx, 12345)

	assert.NoError(t, err)
	assert.Len(t, result.Comments, 2)
	contentRepo.AssertExpectations(t)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_GetComments_ContentNotFound(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	contentRepo.On("GetByKinopubItemID", ctx, int64(12345)).Return(nil, domain.ErrContentNotFound)

	result, err := service.GetComments(ctx, 12345)

	assert.NoError(t, err)
	assert.Empty(t, result.Comments)
	contentRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_Success(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	user := &domain.User{
		ID:              "user-id",
		DisplayName: "testuser",
		IsBanned:        false,
	}

	content := &domain.Content{
		ID:            "content-id",
		KinopubItemID: 12345,
	}

	req := &domain.CreateCommentRequest{
		Text:    "test comment",
		Spoiler: false,
		UserID:  "user-id",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(user, nil)
	contentRepo.On("GetOrCreate", ctx, int64(12345)).Return(content, nil)
	commentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Comment")).Return(nil)

	comment, err := service.CreateComment(ctx, 12345, "user-id", req)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "test comment", comment.Text)
	assert.Equal(t, "content-id", comment.ContentID)
	userRepo.AssertExpectations(t)
	contentRepo.AssertExpectations(t)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_UserBanned(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	user := &domain.User{
		ID:       "user-id",
		IsBanned: true,
	}

	req := &domain.CreateCommentRequest{
		Text:   "test comment",
		UserID: "user-id",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(user, nil)

	comment, err := service.CreateComment(ctx, 12345, "user-id", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserBanned, err)
	assert.Nil(t, comment)
	userRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_EmptyText(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	req := &domain.CreateCommentRequest{
		Text:   "",
		UserID: "user-id",
	}

	comment, err := service.CreateComment(ctx, 12345, "user-id", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentEmpty, err)
	assert.Nil(t, comment)
}

func TestCommentService_ReplyToComment_Success(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	user := &domain.User{
		ID:       "user-id",
		IsBanned: false,
	}

	parentComment := &domain.Comment{
		ID:        "parent-id",
		ContentID: "content-id",
		Text:      "parent comment",
		ParentID:  nil,
	}

	req := &domain.CreateCommentRequest{
		Text:   "reply comment",
		UserID: "user-id",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(user, nil)
	commentRepo.On("GetByID", ctx, "parent-id").Return(parentComment, nil)
	commentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Comment")).Return(nil)

	reply, err := service.ReplyToComment(ctx, "parent-id", "user-id", req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, "reply comment", reply.Text)
	assert.Equal(t, "parent-id", *reply.ParentID)
	userRepo.AssertExpectations(t)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_ReplyToComment_ParentDeleted(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	user := &domain.User{
		ID:       "user-id",
		IsBanned: false,
	}

	deletedAt := time.Now()
	parentComment := &domain.Comment{
		ID:        "parent-id",
		ContentID: "content-id",
		DeletedAt: &deletedAt,
	}

	req := &domain.CreateCommentRequest{
		Text:   "reply comment",
		UserID: "user-id",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(user, nil)
	commentRepo.On("GetByID", ctx, "parent-id").Return(parentComment, nil)

	reply, err := service.ReplyToComment(ctx, "parent-id", "user-id", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentDeleted, err)
	assert.Nil(t, reply)
}

func TestCommentService_UpdateComment_Success(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	existingComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "user-id",
		Text:      "original text",
	}

	spoiler := true
	req := &domain.UpdateCommentRequest{
		Text:    "updated text",
		Spoiler: &spoiler,
		UserID:  "user-id",
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(existingComment, nil)
	commentRepo.On("Update", ctx, mock.AnythingOfType("*domain.Comment")).Return(nil)

	comment, err := service.UpdateComment(ctx, "comment-id", "user-id", req)

	assert.NoError(t, err)
	assert.Equal(t, "updated text", comment.Text)
	assert.True(t, comment.Spoiler)
	assert.NotNil(t, comment.EditedAt)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_UpdateComment_NotOwner(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	existingComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "owner-id",
		Text:      "original text",
	}

	req := &domain.UpdateCommentRequest{
		Text:   "updated text",
		UserID: "other-user",
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(existingComment, nil)

	comment, err := service.UpdateComment(ctx, "comment-id", "other-user", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotCommentOwner, err)
	assert.Nil(t, comment)
}

func TestCommentService_DeleteComment_Success(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	existingComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "user-id",
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(existingComment, nil)
	commentRepo.On("SoftDelete", ctx, "comment-id").Return(nil)

	err := service.DeleteComment(ctx, "comment-id", "user-id")

	assert.NoError(t, err)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_DeleteComment_NotOwner(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	existingComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "owner-id",
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(existingComment, nil)

	err := service.DeleteComment(ctx, "comment-id", "other-user")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotCommentOwner, err)
}

func TestCommentService_AdminDelete_Success(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	existingComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "user-id",
	}

	deletedAt := time.Now()
	deletedComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "user-id",
		DeletedAt: &deletedAt,
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(existingComment, nil).Once()
	commentRepo.On("SoftDelete", ctx, "comment-id").Return(nil)
	commentRepo.On("GetByID", ctx, "comment-id").Return(deletedComment, nil).Once()

	comment, err := service.AdminDelete(ctx, "comment-id")

	assert.NoError(t, err)
	assert.True(t, comment.IsDeleted())
	commentRepo.AssertExpectations(t)
}

func TestCommentService_GetContentByID(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	content := &domain.Content{
		ID:            "content-id",
		KinopubItemID: 12345,
	}

	contentRepo.On("GetByID", ctx, "content-id").Return(content, nil)

	result, err := service.GetContentByID(ctx, "content-id")

	assert.NoError(t, err)
	assert.Equal(t, int64(12345), result.KinopubItemID)
	contentRepo.AssertExpectations(t)
}

func TestCommentService_GetByID(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	comment := &domain.Comment{
		ID:   "comment-id",
		Text: "test comment",
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(comment, nil)

	result, err := service.GetByID(ctx, "comment-id")

	assert.NoError(t, err)
	assert.Equal(t, "test comment", result.Text)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_GetByID_NotFound(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	commentRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrCommentNotFound)

	_, err := service.GetByID(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentNotFound, err)
}

func TestCommentService_List(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	comments := []domain.Comment{
		{ID: "c1", Text: "comment 1"},
		{ID: "c2", Text: "comment 2"},
	}

	commentRepo.On("List", ctx, 0, 20, false).Return(comments, 2, nil)

	result, total, err := service.List(ctx, 1, 20, false)

	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
	commentRepo.AssertExpectations(t)
}

func TestCommentService_CreateComment_UserNotFound(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	req := &domain.CreateCommentRequest{
		Text:   "test comment",
		UserID: "nonexistent",
	}

	userRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrUserNotFound)

	_, err := service.CreateComment(ctx, 12345, "nonexistent", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserNotFound, err)
}

func TestCommentService_ReplyToComment_UserBanned(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	user := &domain.User{
		ID:       "user-id",
		IsBanned: true,
	}

	req := &domain.CreateCommentRequest{
		Text:   "reply comment",
		UserID: "user-id",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(user, nil)

	_, err := service.ReplyToComment(ctx, "parent-id", "user-id", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserBanned, err)
}

func TestCommentService_ReplyToComment_ParentNotFound(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	user := &domain.User{
		ID:       "user-id",
		IsBanned: false,
	}

	req := &domain.CreateCommentRequest{
		Text:   "reply comment",
		UserID: "user-id",
	}

	userRepo.On("GetByID", ctx, "user-id").Return(user, nil)
	commentRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrCommentNotFound)

	_, err := service.ReplyToComment(ctx, "nonexistent", "user-id", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentNotFound, err)
}

func TestCommentService_UpdateComment_NotFound(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	req := &domain.UpdateCommentRequest{
		Text:   "updated",
		UserID: "user-id",
	}

	commentRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrCommentNotFound)

	_, err := service.UpdateComment(ctx, "nonexistent", "user-id", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentNotFound, err)
}

func TestCommentService_UpdateComment_AlreadyDeleted(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	deletedAt := time.Now()
	comment := &domain.Comment{
		ID:        "comment-id",
		UserID:    "user-id",
		DeletedAt: &deletedAt,
	}

	req := &domain.UpdateCommentRequest{
		Text:   "updated",
		UserID: "user-id",
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(comment, nil)

	_, err := service.UpdateComment(ctx, "comment-id", "user-id", req)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentDeleted, err)
}

func TestCommentService_DeleteComment_NotFound(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	commentRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrCommentNotFound)

	err := service.DeleteComment(ctx, "nonexistent", "user-id")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentNotFound, err)
}

func TestCommentService_DeleteComment_AlreadyDeleted(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	deletedAt := time.Now()
	comment := &domain.Comment{
		ID:        "comment-id",
		UserID:    "user-id",
		DeletedAt: &deletedAt,
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(comment, nil)

	err := service.DeleteComment(ctx, "comment-id", "user-id")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentDeleted, err)
}

func TestCommentService_AdminDelete_NotFound(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	commentRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrCommentNotFound)

	_, err := service.AdminDelete(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCommentNotFound, err)
}

func TestCommentService_AdminDelete_Restore(t *testing.T) {
	commentRepo := new(mocks.MockCommentRepository)
	contentRepo := new(mocks.MockContentRepository)
	userRepo := new(mocks.MockUserRepository)
	service := NewCommentService(commentRepo, contentRepo, userRepo)
	ctx := context.Background()

	deletedAt := time.Now()
	deletedComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "user-id",
		DeletedAt: &deletedAt,
	}

	restoredComment := &domain.Comment{
		ID:        "comment-id",
		ContentID: "content-id",
		UserID:    "user-id",
	}

	commentRepo.On("GetByID", ctx, "comment-id").Return(deletedComment, nil).Once()
	commentRepo.On("Restore", ctx, "comment-id").Return(nil)
	commentRepo.On("GetByID", ctx, "comment-id").Return(restoredComment, nil).Once()

	comment, err := service.AdminDelete(ctx, "comment-id")

	assert.NoError(t, err)
	assert.False(t, comment.IsDeleted())
	commentRepo.AssertExpectations(t)
}
