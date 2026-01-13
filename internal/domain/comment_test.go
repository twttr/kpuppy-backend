package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateCommentRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateCommentRequest
		wantErr error
	}{
		{
			name:    "valid comment",
			req:     CreateCommentRequest{Text: "Hello world"},
			wantErr: nil,
		},
		{
			name:    "empty text",
			req:     CreateCommentRequest{Text: ""},
			wantErr: ErrCommentEmpty,
		},
		{
			name:    "text too long",
			req:     CreateCommentRequest{Text: strings.Repeat("a", MaxCommentLength+1)},
			wantErr: ErrCommentTooLong,
		},
		{
			name:    "exactly max length",
			req:     CreateCommentRequest{Text: strings.Repeat("a", MaxCommentLength)},
			wantErr: nil,
		},
		{
			name:    "unicode characters within limit",
			req:     CreateCommentRequest{Text: strings.Repeat("日", 500)},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestComment_ToResponse(t *testing.T) {
	now := time.Now()
	editedAt := now.Add(time.Hour)

	user := &User{
		ID:              "user-123",
		DisplayName: "testuser",
		Avatar:          strPtr("https://example.com/avatar.jpg"),
	}

	comment := &Comment{
		ID:        "comment-123",
		ContentID: "content-456",
		UserID:    "user-123",
		User:      user,
		Text:      "Test comment",
		Spoiler:   true,
		ParentID:  nil,
		EditedAt:  &editedAt,
		DeletedAt: nil,
		CreatedAt: now,
		Replies:   []Comment{},
	}

	resp := comment.ToResponse()

	assert.Equal(t, "comment-123", resp.ID)
	assert.Equal(t, "content-456", resp.ContentID)
	assert.Equal(t, "user-123", resp.UserID)
	assert.Equal(t, "testuser", resp.User.DisplayName)
	assert.Equal(t, "Test comment", resp.Text)
	assert.True(t, resp.Spoiler)
	assert.Nil(t, resp.ParentID)
	assert.NotNil(t, resp.EditedAt)
	assert.Equal(t, editedAt.Unix(), *resp.EditedAt)
	assert.Equal(t, now.Unix(), resp.CreatedAt)
}

func TestComment_ToResponse_Deleted(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(time.Hour)

	user := &User{
		ID:              "user-123",
		DisplayName: "testuser",
	}

	comment := &Comment{
		ID:        "comment-123",
		ContentID: "content-456",
		UserID:    "user-123",
		User:      user,
		Text:      "Original text",
		DeletedAt: &deletedAt,
		CreatedAt: now,
	}

	resp := comment.ToResponse()

	assert.Equal(t, "[deleted]", resp.Text)
}

func TestComment_IsDeleted(t *testing.T) {
	now := time.Now()

	comment := &Comment{DeletedAt: nil}
	assert.False(t, comment.IsDeleted())

	comment.DeletedAt = &now
	assert.True(t, comment.IsDeleted())
}

func TestUpdateCommentRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateCommentRequest
		wantErr error
	}{
		{
			name:    "valid update",
			req:     UpdateCommentRequest{Text: "Updated text"},
			wantErr: nil,
		},
		{
			name:    "empty text",
			req:     UpdateCommentRequest{Text: ""},
			wantErr: ErrCommentEmpty,
		},
		{
			name:    "text too long",
			req:     UpdateCommentRequest{Text: strings.Repeat("a", MaxCommentLength+1)},
			wantErr: ErrCommentTooLong,
		},
		{
			name:    "exactly max length",
			req:     UpdateCommentRequest{Text: strings.Repeat("a", MaxCommentLength)},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestComment_ToResponse_WithReplies(t *testing.T) {
	now := time.Now()
	parentID := "parent-123"

	parent := &Comment{
		ID:        "parent-123",
		ContentID: "content-456",
		UserID:    "user-123",
		User:      &User{ID: "user-123", DisplayName: "parent_user"},
		Text:      "Parent comment",
		CreatedAt: now,
		Replies: []Comment{
			{
				ID:        "reply-1",
				ContentID: "content-456",
				UserID:    "user-456",
				User:      &User{ID: "user-456", DisplayName: "reply_user"},
				Text:      "Reply 1",
				ParentID:  &parentID,
				CreatedAt: now,
			},
		},
	}

	resp := parent.ToResponse()

	assert.Len(t, resp.Replies, 1)
	assert.Equal(t, "reply-1", resp.Replies[0].ID)
	assert.Equal(t, "Reply 1", resp.Replies[0].Text)
}

func TestComment_ToResponse_NilUser(t *testing.T) {
	now := time.Now()

	comment := &Comment{
		ID:        "comment-123",
		ContentID: "content-456",
		UserID:    "user-123",
		User:      nil,
		Text:      "Test comment",
		CreatedAt: now,
	}

	resp := comment.ToResponse()

	assert.Equal(t, UserResponse{}, resp.User)
}

func TestComment_ToResponse_NoEditedAt(t *testing.T) {
	now := time.Now()

	comment := &Comment{
		ID:        "comment-123",
		ContentID: "content-456",
		UserID:    "user-123",
		User:      &User{ID: "user-123", DisplayName: "testuser"},
		Text:      "Test comment",
		EditedAt:  nil,
		CreatedAt: now,
	}

	resp := comment.ToResponse()

	assert.Nil(t, resp.EditedAt)
}

func strPtr(s string) *string {
	return &s
}
