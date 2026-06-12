package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type salonService struct {
	salonRepo repository.SalonRepository
	userRepo  repository.UserRepository
}

func NewSalonService(salonRepo repository.SalonRepository, userRepo repository.UserRepository) service.SalonService {
	return &salonService{salonRepo: salonRepo, userRepo: userRepo}
}

func (s *salonService) RegisterSalon(ctx context.Context, userID string, req service.RegisterSalonRequest) (*domain.Salon, error) {
	// 1ユーザー1サロン制
	if _, err := s.salonRepo.FindByOwnerID(ctx, userID); err == nil {
		return nil, domain.ErrInvalidRequest
	}

	salon := &domain.Salon{
		ID:                 uuid.NewString(),
		OwnerID:            userID,
		Name:               req.Name,
		Description:        req.Description,
		Address:            req.Address,
		Prefecture:         req.Prefecture,
		City:               req.City,
		Lat:                req.Lat,
		Lng:                req.Lng,
		Phone:              req.Phone,
		SkillBadgeTags:     []string{},
		PortfolioImageURLs: []string{},
		VerificationStatus: "pending",
		IsActive:           false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if err := s.salonRepo.Create(ctx, salon); err != nil {
		return nil, err
	}
	return salon, nil
}

func (s *salonService) ListSalons(ctx context.Context, filter service.SalonFilter) ([]*domain.Salon, string, error) {
	return s.salonRepo.List(ctx, filter.Prefecture, filter.VerificationStatus, filter.SkillBadgeTags, filter.Sort, filter.Cursor, filter.Limit)
}

func (s *salonService) GetSalon(ctx context.Context, id string) (*domain.Salon, error) {
	return s.salonRepo.FindByID(ctx, id)
}

func (s *salonService) UpdateSalon(ctx context.Context, ownerID, salonID string, req service.UpdateSalonRequest) (*domain.Salon, error) {
	salon, err := s.salonRepo.FindByID(ctx, salonID)
	if err != nil {
		return nil, err
	}
	if salon.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	if req.Name != nil {
		salon.Name = *req.Name
	}
	if req.Description != nil {
		salon.Description = req.Description
	}
	if req.Address != nil {
		salon.Address = *req.Address
	}
	if req.Phone != nil {
		salon.Phone = req.Phone
	}
	salon.UpdatedAt = time.Now()
	if err := s.salonRepo.Update(ctx, salon); err != nil {
		return nil, err
	}
	return salon, nil
}

func (s *salonService) AddPortfolio(ctx context.Context, ownerID, salonID, imageURL string) error {
	salon, err := s.salonRepo.FindByID(ctx, salonID)
	if err != nil {
		return err
	}
	if salon.OwnerID != ownerID {
		return domain.ErrForbidden
	}
	salon.PortfolioImageURLs = append(salon.PortfolioImageURLs, imageURL)
	salon.UpdatedAt = time.Now()
	return s.salonRepo.Update(ctx, salon)
}

func (s *salonService) VerifySalon(ctx context.Context, adminID, salonID, action string, reason *string) error {
	salon, err := s.salonRepo.FindByID(ctx, salonID)
	if err != nil {
		return err
	}
	switch action {
	case "approve":
		salon.VerificationStatus = "approved"
		salon.IsActive = true
	case "reject":
		salon.VerificationStatus = "rejected"
		salon.IsActive = false
	default:
		return domain.ErrInvalidRequest
	}
	salon.UpdatedAt = time.Now()
	return s.salonRepo.Update(ctx, salon)
}
