package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/twttr/kpuppy-backend/internal/domain"
	"github.com/twttr/kpuppy-backend/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Provision(ctx context.Context, req *domain.ProvisionRequest) (*domain.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.userRepo.GetByUserHash(ctx, req.UserHash)
	if err == nil {
		if req.Avatar != nil && (existing.Avatar == nil || *existing.Avatar != *req.Avatar) {
			_ = s.userRepo.UpdateAvatar(ctx, existing.ID, req.Avatar)
			existing.Avatar = req.Avatar
		}
		return existing, nil
	}
	if err != domain.ErrUserNotFound {
		return nil, err
	}

	displayName := domain.GeneratePseudonym(req.UserHash)

	user := &domain.User{
		ID:          uuid.New().String(),
		UserHash:    req.UserHash,
		DisplayName: displayName,
		Avatar:      req.Avatar,
		IsBanned:    false,
		CreatedAt:   time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, page, perPage int) ([]domain.User, int, error) {
	offset := (page - 1) * perPage
	return s.userRepo.List(ctx, offset, perPage)
}

func (s *UserService) SetBanned(ctx context.Context, id string, banned bool) error {
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.userRepo.SetBanned(ctx, id, banned)
}
