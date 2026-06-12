package impl

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type auctionService struct {
	auctionRepo repository.AuctionRepository
	salonRepo   repository.SalonRepository
	sqsClient   *sqs.Client
	sqsQueueURL string
}

func NewAuctionService(
	auctionRepo repository.AuctionRepository,
	salonRepo repository.SalonRepository,
	sqsClient *sqs.Client,
	sqsQueueURL string,
) service.AuctionService {
	return &auctionService{
		auctionRepo: auctionRepo,
		salonRepo:   salonRepo,
		sqsClient:   sqsClient,
		sqsQueueURL: sqsQueueURL,
	}
}

func (s *auctionService) CreateBookingRequest(ctx context.Context, userID string, req service.CreateBookingRequestInput) (*domain.BookingRequest, error) {
	// 同時オープンリクエストは3件まで
	count, err := s.auctionRepo.CountOpenRequestsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 3 {
		return nil, domain.ErrRequestLimitReached
	}

	from, err := time.Parse(time.RFC3339, req.DesiredDateFrom)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	to, err := time.Parse(time.RFC3339, req.DesiredDateTo)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}

	br := &domain.BookingRequest{
		ID:               uuid.NewString(),
		UserID:           userID,
		DesignIPID:       req.DesignIPID,
		ARSessionID:      req.ARSessionID,
		NailDataSnapshot: req.NailDataSnapshot,
		BudgetMaxYen:     req.BudgetMaxYen,
		DesiredDateFrom:  from,
		DesiredDateTo:    to,
		AreaPrefecture:   req.AreaPrefecture,
		AreaCity:         req.AreaCity,
		Status:           domain.BookingRequestOpen,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
	}
	if err := s.auctionRepo.CreateBookingRequest(ctx, br); err != nil {
		return nil, err
	}
	return br, nil
}

func (s *auctionService) GetBookingRequest(ctx context.Context, userID, requestID string) (*domain.BookingRequest, []*domain.Bid, error) {
	req, err := s.auctionRepo.FindBookingRequestByID(ctx, requestID)
	if err != nil {
		return nil, nil, err
	}
	// 本人か管理者のみ全入札を閲覧可（サロンは自分の入札のみ）
	bids, err := s.auctionRepo.ListBidsByRequestID(ctx, requestID)
	if err != nil {
		return nil, nil, err
	}
	return req, bids, nil
}

func (s *auctionService) CancelBookingRequest(ctx context.Context, userID, requestID string) error {
	req, err := s.auctionRepo.FindBookingRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req.UserID != userID {
		return domain.ErrForbidden
	}
	if req.Status != domain.BookingRequestOpen && req.Status != domain.BookingRequestBidding {
		return domain.ErrInvalidRequest
	}
	return s.auctionRepo.UpdateBookingRequestStatus(ctx, requestID, domain.BookingRequestCancelled)
}

func (s *auctionService) ListMatchingRequests(ctx context.Context, salonID string, cursor string, limit int) ([]*domain.BookingRequest, string, error) {
	salon, err := s.salonRepo.FindByOwnerID(ctx, salonID)
	if err != nil {
		return nil, "", err
	}
	return s.auctionRepo.ListOpenRequestsForSalon(ctx, salon.Prefecture, cursor, limit)
}

func (s *auctionService) PlaceBid(ctx context.Context, userID, requestID string, req service.PlaceBidInput) (*domain.Bid, error) {
	// ログイン中ユーザーのサロンを解決する（FK: bids.salon_id → salons.id）
	salon, err := s.salonRepo.FindByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	salonID := salon.ID

	bookingReq, err := s.auctionRepo.FindBookingRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if bookingReq.Status != domain.BookingRequestOpen {
		return nil, domain.ErrRequestExpired
	}
	if req.PriceYen > bookingReq.BudgetMaxYen {
		return nil, domain.ErrBudgetExceeded
	}

	// 重複入札チェック
	if _, err := s.auctionRepo.FindBidBySalonAndRequest(ctx, salonID, requestID); err == nil {
		return nil, domain.ErrAlreadyBid
	}

	slotAt, err := time.Parse(time.RFC3339, req.AvailableSlotAt)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}

	bid := &domain.Bid{
		ID:                    uuid.NewString(),
		BookingRequestID:      requestID,
		SalonID:               salonID,
		PriceYen:              req.PriceYen,
		IncludesRemoval:       req.IncludesRemoval,
		RemovalFeeYen:         req.RemovalFeeYen,
		AvailableSlotAt:       slotAt,
		DynamicDiscountReason: req.DynamicDiscountReason,
		Message:               req.Message,
		Status:                domain.BidPending,
		ExpiresAt:             bookingReq.ExpiresAt,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	if err := s.auctionRepo.CreateBid(ctx, bid); err != nil {
		return nil, err
	}
	return bid, nil
}

func (s *auctionService) UpdateBid(ctx context.Context, userID, requestID, bidID string, newPriceYen int) (*domain.Bid, error) {
	// ログイン中ユーザーのサロンを解決する
	salon, err := s.salonRepo.FindByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	salonID := salon.ID

	bid, err := s.auctionRepo.FindBidByID(ctx, bidID)
	if err != nil {
		return nil, err
	}
	if bid.SalonID != salonID || bid.BookingRequestID != requestID {
		return nil, domain.ErrForbidden
	}
	if bid.RebidCount >= 1 {
		return nil, domain.ErrRebidLimitReached
	}
	if newPriceYen >= bid.PriceYen {
		return nil, domain.ErrInvalidRequest // 再入札は値下げのみ
	}
	bid.PriceYen = newPriceYen
	bid.UpdatedAt = time.Now()
	if err := s.auctionRepo.UpdateBid(ctx, bid); err != nil {
		return nil, err
	}
	return bid, nil
}

func (s *auctionService) ListBids(ctx context.Context, userID, requestID, sort string) ([]*domain.Bid, error) {
	return s.auctionRepo.ListBidsByRequestID(ctx, requestID)
}
