package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	bookingRepo repository.BookingRepository
	auctionRepo repository.AuctionRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository, bookingRepo repository.BookingRepository, auctionRepo repository.AuctionRepository) service.ReviewService {
	return &reviewService{reviewRepo: reviewRepo, bookingRepo: bookingRepo, auctionRepo: auctionRepo}
}

func (s *reviewService) PostReview(ctx context.Context, userID string, req service.PostReviewInput) (*domain.Review, error) {
	booking, err := s.bookingRepo.FindByID(ctx, req.BookingID)
	if err != nil {
		return nil, err
	}
	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if booking.Status != domain.BookingCompleted {
		return nil, domain.ErrInvalidRequest
	}
	// 重複レビューチェック
	if _, err := s.reviewRepo.FindByBookingID(ctx, req.BookingID); err == nil {
		return nil, domain.ErrInvalidRequest
	}

	// DesignIPID を booking → booking_request チェーンから取得する
	bookingReq, err := s.auctionRepo.FindBookingRequestByID(ctx, booking.BookingRequestID)
	if err != nil {
		return nil, err
	}

	review := &domain.Review{
		ID:                uuid.NewString(),
		BookingID:         req.BookingID,
		UserID:            userID,
		SalonID:           booking.SalonID,
		DesignIPID:        bookingReq.DesignIPID,
		ReproductionScore: req.ReproductionScore,
		OverallScore:      req.OverallScore,
		Comment:           req.Comment,
		BeforePhotoURL:    req.BeforePhotoURL,
		AfterPhotoURL:     req.AfterPhotoURL,
		AIAnalysisStatus:  "pending",
		CreatedAt:         time.Now(),
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (s *reviewService) ListSalonReviews(ctx context.Context, salonID string, cursor string, limit int) ([]*domain.Review, string, error) {
	return s.reviewRepo.ListBySalonID(ctx, salonID, cursor, limit)
}
