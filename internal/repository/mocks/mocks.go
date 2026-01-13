package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/twttr/kpuppy-backend/internal/domain"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUserHash(ctx context.Context, userHash string) (*domain.User, error) {
	args := m.Called(ctx, userHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateAvatar(ctx context.Context, id string, avatar *string) error {
	args := m.Called(ctx, id, avatar)
	return args.Error(0)
}

func (m *MockUserRepository) SetBanned(ctx context.Context, id string, banned bool) error {
	args := m.Called(ctx, id, banned)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]domain.User, int, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

type MockContentRepository struct {
	mock.Mock
}

func (m *MockContentRepository) GetByID(ctx context.Context, id string) (*domain.Content, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Content), args.Error(1)
}

func (m *MockContentRepository) GetByKinopubItemID(ctx context.Context, kinopubItemID int64) (*domain.Content, error) {
	args := m.Called(ctx, kinopubItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Content), args.Error(1)
}

func (m *MockContentRepository) Create(ctx context.Context, content *domain.Content) error {
	args := m.Called(ctx, content)
	return args.Error(0)
}

func (m *MockContentRepository) GetOrCreate(ctx context.Context, kinopubItemID int64) (*domain.Content, error) {
	args := m.Called(ctx, kinopubItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Content), args.Error(1)
}

type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) GetByID(ctx context.Context, id string) (*domain.Comment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetByContentID(ctx context.Context, contentID string) ([]domain.Comment, error) {
	args := m.Called(ctx, contentID)
	return args.Get(0).([]domain.Comment), args.Error(1)
}

func (m *MockCommentRepository) GetReplies(ctx context.Context, parentID string) ([]domain.Comment, error) {
	args := m.Called(ctx, parentID)
	return args.Get(0).([]domain.Comment), args.Error(1)
}

func (m *MockCommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	args := m.Called(ctx, comment)
	return args.Error(0)
}

func (m *MockCommentRepository) SoftDelete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) Restore(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCommentRepository) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]domain.Comment, int, error) {
	args := m.Called(ctx, offset, limit, includeDeleted)
	return args.Get(0).([]domain.Comment), args.Int(1), args.Error(2)
}
