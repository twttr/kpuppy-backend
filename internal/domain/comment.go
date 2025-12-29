package domain

import (
	"time"
	"unicode/utf8"
)

const MaxCommentLength = 1000

type Comment struct {
	ID            string     `json:"id"`
	ContentID     string     `json:"contentId"`
	KinopubItemID int64      `json:"-"`
	UserID        string     `json:"userId"`
	User          *User      `json:"-"`
	Text          string     `json:"text"`
	Spoiler       bool       `json:"spoiler"`
	ParentID      *string    `json:"parentId"`
	EditedAt      *time.Time `json:"editedAt,omitempty"`
	DeletedAt     *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"createdAt"`
	Replies       []Comment  `json:"replies,omitempty"`
}

type CommentResponse struct {
	ID        string            `json:"id"`
	ContentID string            `json:"contentId"`
	UserID    string            `json:"userId"`
	User      UserResponse      `json:"user"`
	Text      string            `json:"text"`
	Spoiler   bool              `json:"spoiler"`
	ParentID  *string           `json:"parentId"`
	EditedAt  *int64            `json:"editedAt,omitempty"`
	CreatedAt int64             `json:"createdAt"`
	Replies   []CommentResponse `json:"replies"`
}

func (c *Comment) ToResponse() CommentResponse {
	var editedAt *int64
	if c.EditedAt != nil {
		ts := c.EditedAt.Unix()
		editedAt = &ts
	}

	text := c.Text
	if c.DeletedAt != nil {
		text = "[deleted]"
	}

	var userResp UserResponse
	if c.User != nil {
		userResp = c.User.ToResponse()
	}

	replies := make([]CommentResponse, 0, len(c.Replies))
	for _, reply := range c.Replies {
		replies = append(replies, reply.ToResponse())
	}

	return CommentResponse{
		ID:        c.ID,
		ContentID: c.ContentID,
		UserID:    c.UserID,
		User:      userResp,
		Text:      text,
		Spoiler:   c.Spoiler,
		ParentID:  c.ParentID,
		EditedAt:  editedAt,
		CreatedAt: c.CreatedAt.Unix(),
		Replies:   replies,
	}
}

func (c *Comment) IsDeleted() bool {
	return c.DeletedAt != nil
}

type CreateCommentRequest struct {
	UserID  string `json:"-"`
	Text    string `json:"text"`
	Spoiler bool   `json:"spoiler"`
}

func (r *CreateCommentRequest) Validate() error {
	if r.Text == "" {
		return ErrCommentEmpty
	}
	if utf8.RuneCountInString(r.Text) > MaxCommentLength {
		return ErrCommentTooLong
	}
	return nil
}

type UpdateCommentRequest struct {
	UserID  string `json:"-"`
	Text    string `json:"text"`
	Spoiler *bool  `json:"spoiler,omitempty"`
}

func (r *UpdateCommentRequest) Validate() error {
	if r.Text == "" {
		return ErrCommentEmpty
	}
	if utf8.RuneCountInString(r.Text) > MaxCommentLength {
		return ErrCommentTooLong
	}
	return nil
}

type CommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
}

type PaginationResponse struct {
	Page       int  `json:"page"`
	TotalPages int  `json:"totalPages"`
	TotalItems int  `json:"totalItems"`
	HasMore    bool `json:"hasMore"`
}
