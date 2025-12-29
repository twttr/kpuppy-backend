package repository

import (
	"context"

	"github.com/twttr/kpuppy-backend/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByKinopubUsername(ctx context.Context, username string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	UpdateAvatar(ctx context.Context, id string, avatar *string) error
	SetBanned(ctx context.Context, id string, banned bool) error
	List(ctx context.Context, offset, limit int) ([]domain.User, int, error)
}

type ContentRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Content, error)
	GetByKinopubItemID(ctx context.Context, kinopubItemID int64) (*domain.Content, error)
	Create(ctx context.Context, content *domain.Content) error
	GetOrCreate(ctx context.Context, kinopubItemID int64) (*domain.Content, error)
}

type CommentRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Comment, error)
	GetByContentID(ctx context.Context, contentID string) ([]domain.Comment, error)
	GetReplies(ctx context.Context, parentID string) ([]domain.Comment, error)
	Create(ctx context.Context, comment *domain.Comment) error
	Update(ctx context.Context, comment *domain.Comment) error
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int, includeDeleted bool) ([]domain.Comment, int, error)
}
