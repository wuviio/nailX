package service

import (
	"context"

	"github.com/nailx/backend/internal/domain"
)

// --- Auth / User ---

type AuthService interface {
	Register(ctx context.Context, firebaseToken, displayName string, gender *string) (*domain.User, error)
}

type UserService interface {
	GetMe(ctx context.Context, userID string) (*domain.User, *domain.NailProfile, error)
	UpdateMe(ctx context.Context, userID string, req UpdateUserRequest) (*domain.User, error)
	GetPublicProfile(ctx context.Context, targetID string) (*domain.User, error)
	UpsertNailProfile(ctx context.Context, userID string, req UpsertNailProfileRequest) (*domain.NailProfile, error)
}

type UpdateUserRequest struct {
	DisplayName   *string
	AvatarURL     *string
	Bio           *string
	LifestyleTags []string
}

type UpsertNailProfileRequest struct {
	AvgNailLengthMM *float64
	NailShape       *string
	GelLiftTendency *string
	AllergyNotes    *string
}

// --- AR ---

type ARService interface {
	CreateSession(ctx context.Context, userID string, req CreateARSessionRequest) (*domain.ARSession, error)
	GetSession(ctx context.Context, userID, sessionID string) (*domain.ARSession, error)
}

type CreateARSessionRequest struct {
	DesignIPID           *string
	DetectedNailLengthMM *float64
	HasExistingGel       *bool
	DetectedNailShape    *string
	HandSnapshotURL      *string
}

// --- Design IP ---

type DesignService interface {
	RegisterDesign(ctx context.Context, userID string, req RegisterDesignRequest) (designID, jobID string, err error)
	GetSimilarityCheckStatus(ctx context.Context, jobID string) (*SimilarityCheckResult, error)
	ListFeed(ctx context.Context, filter DesignFeedFilter) ([]*domain.DesignIP, string, error)
	GetDesign(ctx context.Context, id string) (*domain.DesignIP, []*domain.DesignRoyaltyNode, error)
	ListByCreator(ctx context.Context, creatorID string, cursor string, limit int) ([]*domain.DesignIP, string, error)
	UpdateDesign(ctx context.Context, userID string, callerRole domain.UserRole, designID string, req UpdateDesignRequest) (*domain.DesignIP, error)
}

type RegisterDesignRequest struct {
	Title           string
	Description     *string
	PreviewImageURL string
	DesignData      []byte
	ParentIPID      *string
	GenderTag       string
	StyleTags       []string
	IsPublic        bool
}

type SimilarityCheckResult struct {
	Status       string
	Result       *domain.SimilarityResult
	SimilarIPID  *string
	Score        *float64
}

type DesignFeedFilter struct {
	GenderTag *string
	StyleTags []string
	Status    *string
	Sort      string
	Cursor    string
	Limit     int
}

type UpdateDesignRequest struct {
	Title       *string
	Description *string
	IsPublic    *bool
}

// --- Salon ---

type SalonService interface {
	RegisterSalon(ctx context.Context, userID string, req RegisterSalonRequest) (*domain.Salon, error)
	ListSalons(ctx context.Context, filter SalonFilter) ([]*domain.Salon, string, error)
	GetSalon(ctx context.Context, id string) (*domain.Salon, error)
	UpdateSalon(ctx context.Context, ownerID, salonID string, req UpdateSalonRequest) (*domain.Salon, error)
	AddPortfolio(ctx context.Context, ownerID, salonID, imageURL string) error
	VerifySalon(ctx context.Context, adminID, salonID, action string, reason *string) error
}

type RegisterSalonRequest struct {
	Name          string
	Description   *string
	Address       string
	Prefecture    string
	City          *string
	Lat           *float64
	Lng           *float64
	Phone         *string
	BusinessHours []byte
}

type SalonFilter struct {
	Prefecture         *string
	VerificationStatus *string
	SkillBadgeTags     []string
	Sort               string
	Cursor             string
	Limit              int
}

type UpdateSalonRequest struct {
	Name          *string
	Description   *string
	Address       *string
	Phone         *string
	BusinessHours []byte
}

// --- Auction ---

type AuctionService interface {
	CreateBookingRequest(ctx context.Context, userID string, req CreateBookingRequestInput) (*domain.BookingRequest, error)
	GetBookingRequest(ctx context.Context, userID, requestID string) (*domain.BookingRequest, []*domain.Bid, error)
	CancelBookingRequest(ctx context.Context, userID, requestID string) error
	ListMatchingRequests(ctx context.Context, salonID string, cursor string, limit int) ([]*domain.BookingRequest, string, error)

	PlaceBid(ctx context.Context, salonID, requestID string, req PlaceBidInput) (*domain.Bid, error)
	UpdateBid(ctx context.Context, salonID, requestID, bidID string, newPriceYen int) (*domain.Bid, error)
	ListBids(ctx context.Context, userID, requestID, sort string) ([]*domain.Bid, error)
}

type CreateBookingRequestInput struct {
	DesignIPID       string
	ARSessionID      *string
	NailDataSnapshot []byte
	BudgetMaxYen     int
	DesiredDateFrom  string
	DesiredDateTo    string
	AreaPrefecture   string
	AreaCity         *string
}

type PlaceBidInput struct {
	PriceYen             int
	IncludesRemoval      bool
	RemovalFeeYen        int
	AvailableSlotAt      string
	DynamicDiscountReason *string
	Message              *string
}

// --- Booking / Payment ---

type BookingService interface {
	ConfirmBooking(ctx context.Context, userID, bidID, paymentMethodID string) (*domain.Booking, *domain.Payment, string, error)
	ListBookings(ctx context.Context, userID, status string) ([]*domain.Booking, error)
	GetBooking(ctx context.Context, userID, bookingID string) (*domain.Booking, error)
	CompleteBooking(ctx context.Context, salonID, bookingID string) error
	CancelBooking(ctx context.Context, requesterID, bookingID, reason string) error
}

// --- Review ---

type ReviewService interface {
	PostReview(ctx context.Context, userID string, req PostReviewInput) (*domain.Review, error)
	ListSalonReviews(ctx context.Context, salonID string, cursor string, limit int) ([]*domain.Review, string, error)
}

type PostReviewInput struct {
	BookingID         string
	ReproductionScore int
	OverallScore      int
	Comment           *string
	BeforePhotoURL    *string
	AfterPhotoURL     *string
}

// --- Notification ---

type NotificationService interface {
	ListNotifications(ctx context.Context, userID string, isRead *bool, cursor string, limit int) ([]*domain.Notification, string, error)
	MarkRead(ctx context.Context, userID, notificationID string) error
	RegisterFCMToken(ctx context.Context, userID, token, platform string) error
}

// --- Media ---

type MediaService interface {
	GeneratePresignedURL(ctx context.Context, userID, fileType, purpose string) (uploadURL, fileURL string, err error)
}
