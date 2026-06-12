package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type bookingService struct {
	bookingRepo repository.BookingRepository
	auctionRepo repository.AuctionRepository
	paymentRepo repository.PaymentRepository
	salonRepo   repository.SalonRepository
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	auctionRepo repository.AuctionRepository,
	paymentRepo repository.PaymentRepository,
	salonRepo repository.SalonRepository,
) service.BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		auctionRepo: auctionRepo,
		paymentRepo: paymentRepo,
		salonRepo:   salonRepo,
	}
}

func (s *bookingService) ConfirmBooking(ctx context.Context, userID, bidID, paymentMethodID string) (*domain.Booking, *domain.Payment, string, error) {
	bid, err := s.auctionRepo.FindBidByID(ctx, bidID)
	if err != nil || bid == nil {
		return nil, nil, "", domain.ErrNotFound
	}
	if bid.Status != domain.BidPending {
		return nil, nil, "", domain.ErrInvalidRequest
	}

	req, err := s.auctionRepo.FindBookingRequestByID(ctx, bid.BookingRequestID)
	if err != nil || req == nil {
		return nil, nil, "", domain.ErrNotFound
	}
	if req.UserID != userID {
		return nil, nil, "", domain.ErrForbidden
	}

	now := time.Now()
	booking := &domain.Booking{
		ID:               uuid.NewString(),
		BookingRequestID: bid.BookingRequestID,
		BidID:            bid.ID,
		UserID:           userID,
		SalonID:          bid.SalonID,
		ScheduledAt:      bid.AvailableSlotAt,
		Status:           domain.BookingConfirmed,
		ConfirmedAt:      now,
	}
	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, nil, "", err
	}

	// 金額計算: 施術代 + オフ料金
	totalYen := bid.PriceYen + bid.RemovalFeeYen
	platformFee, ipFee, salonPayout := domain.CalcPaymentBreakdown(totalYen)

	payment := &domain.Payment{
		ID:                    uuid.NewString(),
		BookingID:             booking.ID,
		TotalAmountYen:        totalYen,
		PlatformFeeYen:        platformFee,
		SalonPayoutYen:        salonPayout,
		DesignRoyaltyTotalYen: ipFee,
		Status:                domain.PaymentAuthorized,
		AuthorizedAt:          now,
	}
	if paymentMethodID != "" {
		payment.PaymentMethod = &paymentMethodID
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, nil, "", err
	}

	// ステータス更新
	_ = s.auctionRepo.UpdateBookingRequestStatus(ctx, bid.BookingRequestID, domain.BookingRequestConfirmed)

	// Stripe client_secret はStripe統合後に実装（Phase 5）
	stripeClientSecret := "pi_placeholder_secret"

	return booking, payment, stripeClientSecret, nil
}

func (s *bookingService) ListBookings(ctx context.Context, userID, status string) ([]*domain.Booking, error) {
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}
	return s.bookingRepo.ListByUserID(ctx, userID, statusPtr)
}

func (s *bookingService) GetBooking(ctx context.Context, userID, bookingID string) (*domain.Booking, error) {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil || booking == nil {
		return nil, domain.ErrNotFound
	}
	// consumerかsalonのどちらかが閲覧可
	if booking.UserID != userID {
		salon, err := s.salonRepo.FindByOwnerID(ctx, userID)
		if err != nil || salon == nil || salon.ID != booking.SalonID {
			return nil, domain.ErrForbidden
		}
	}
	return booking, nil
}

func (s *bookingService) CompleteBooking(ctx context.Context, salonID, bookingID string) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil || booking == nil {
		return domain.ErrNotFound
	}
	if booking.SalonID != salonID {
		return domain.ErrForbidden
	}
	if booking.Status != domain.BookingConfirmed {
		return domain.ErrInvalidRequest
	}
	return s.bookingRepo.UpdateStatus(ctx, bookingID, domain.BookingCompleted)
}

func (s *bookingService) CancelBooking(ctx context.Context, requesterID, bookingID, reason string) error {
	booking, err := s.bookingRepo.FindByID(ctx, bookingID)
	if err != nil || booking == nil {
		return domain.ErrNotFound
	}
	if booking.Status != domain.BookingConfirmed {
		return domain.ErrInvalidRequest
	}

	var newStatus domain.BookingStatus
	if booking.UserID == requesterID {
		newStatus = domain.BookingCancelledByUser
	} else {
		// サロン側のキャンセルか確認
		salon, err := s.salonRepo.FindByOwnerID(ctx, requesterID)
		if err != nil || salon == nil || salon.ID != booking.SalonID {
			return domain.ErrForbidden
		}
		newStatus = domain.BookingCancelledBySalon
	}
	return s.bookingRepo.UpdateStatus(ctx, bookingID, newStatus)
}
