package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) service.UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetMe(ctx context.Context, userID string) (*domain.User, *domain.NailProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	profile, err := s.userRepo.GetNailProfile(ctx, userID)
	if err == domain.ErrNotFound {
		return user, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return user, profile, nil
}

func (s *userService) UpdateMe(ctx context.Context, userID string, req service.UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Bio != nil {
		user.Bio = req.Bio
	}
	if req.LifestyleTags != nil {
		user.LifestyleTags = req.LifestyleTags
	}
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetPublicProfile(ctx context.Context, targetID string) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, targetID)
}

func (s *userService) UpsertNailProfile(ctx context.Context, userID string, req service.UpsertNailProfileRequest) (*domain.NailProfile, error) {
	existing, err := s.userRepo.GetNailProfile(ctx, userID)
	id := uuid.NewString()
	if err == nil {
		id = existing.ID
	}
	profile := &domain.NailProfile{
		ID:              id,
		UserID:          userID,
		AvgNailLengthMM: req.AvgNailLengthMM,
		NailShape:       req.NailShape,
		GelLiftTendency: req.GelLiftTendency,
		AllergyNotes:    req.AllergyNotes,
		UpdatedAt:       time.Now(),
	}
	if err := s.userRepo.UpsertNailProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
