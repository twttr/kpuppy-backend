package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/twttr/kpuppy-backend/internal/domain"
	"github.com/twttr/kpuppy-backend/internal/repository"
)

type CommentService struct {
	commentRepo repository.CommentRepository
	contentRepo repository.ContentRepository
	userRepo    repository.UserRepository
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	contentRepo repository.ContentRepository,
	userRepo repository.UserRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		contentRepo: contentRepo,
		userRepo:    userRepo,
	}
}

func (s *CommentService) GetComments(ctx context.Context, kinopubItemID int64) (*domain.CommentsResponse, error) {
	content, err := s.contentRepo.GetByKinopubItemID(ctx, kinopubItemID)
	if errors.Is(err, domain.ErrContentNotFound) {
		return &domain.CommentsResponse{
			Comments: []domain.CommentResponse{},
		}, nil
	}
	if err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.GetByContentID(ctx, content.ID)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.CommentResponse, 0, len(comments))
	for _, c := range comments {
		responses = append(responses, c.ToResponse())
	}

	return &domain.CommentsResponse{
		Comments: responses,
	}, nil
}

func (s *CommentService) CreateComment(ctx context.Context, kinopubItemID int64, userID string, req *domain.CreateCommentRequest) (*domain.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.IsBanned {
		return nil, domain.ErrUserBanned
	}

	content, err := s.contentRepo.GetOrCreate(ctx, kinopubItemID)
	if err != nil {
		return nil, err
	}

	comment := &domain.Comment{
		ID:        uuid.New().String(),
		ContentID: content.ID,
		UserID:    userID,
		Text:      req.Text,
		Spoiler:   req.Spoiler,
		CreatedAt: time.Now(),
		User:      user,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) ReplyToComment(ctx context.Context, parentCommentID string, userID string, req *domain.CreateCommentRequest) (*domain.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.IsBanned {
		return nil, domain.ErrUserBanned
	}

	parent, err := s.commentRepo.GetByID(ctx, parentCommentID)
	if err != nil {
		return nil, err
	}

	if parent.IsDeleted() {
		return nil, domain.ErrCommentDeleted
	}

	comment := &domain.Comment{
		ID:        uuid.New().String(),
		ContentID: parent.ContentID,
		UserID:    userID,
		Text:      req.Text,
		Spoiler:   req.Spoiler,
		ParentID:  &parentCommentID,
		CreatedAt: time.Now(),
		User:      user,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID string, userID string, req *domain.UpdateCommentRequest) (*domain.Comment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.IsDeleted() {
		return nil, domain.ErrCommentDeleted
	}

	if comment.UserID != userID {
		return nil, domain.ErrNotCommentOwner
	}

	comment.Text = req.Text
	if req.Spoiler != nil {
		comment.Spoiler = *req.Spoiler
	}
	now := time.Now()
	comment.EditedAt = &now

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID string, userID string) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	if comment.IsDeleted() {
		return domain.ErrCommentDeleted
	}

	if comment.UserID != userID {
		return domain.ErrNotCommentOwner
	}

	return s.commentRepo.SoftDelete(ctx, commentID)
}

func (s *CommentService) GetByID(ctx context.Context, commentID string) (*domain.Comment, error) {
	return s.commentRepo.GetByID(ctx, commentID)
}

func (s *CommentService) AdminDelete(ctx context.Context, commentID string) (*domain.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.IsDeleted() {
		if err := s.commentRepo.Restore(ctx, commentID); err != nil {
			return nil, err
		}
	} else {
		if err := s.commentRepo.SoftDelete(ctx, commentID); err != nil {
			return nil, err
		}
	}

	return s.commentRepo.GetByID(ctx, commentID)
}

func (s *CommentService) AdminToggleSpoiler(ctx context.Context, commentID string) (*domain.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	comment.Spoiler = !comment.Spoiler

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) List(ctx context.Context, page, perPage int, includeDeleted bool) ([]domain.Comment, int, error) {
	offset := (page - 1) * perPage
	return s.commentRepo.List(ctx, offset, perPage, includeDeleted)
}

func (s *CommentService) GetContentByID(ctx context.Context, contentID string) (*domain.Content, error) {
	return s.contentRepo.GetByID(ctx, contentID)
}
