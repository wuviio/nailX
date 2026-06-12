package repository

import (
	"context"

	"github.com/nailx/backend/internal/domain"
)

// UserRepository はユーザーデータの永続化を担う
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByFirebaseUID(ctx context.Context, uid string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpsertNailProfile(ctx context.Context, profile *domain.NailProfile) error
	GetNailProfile(ctx context.Context, userID string) (*domain.NailProfile, error)
	UpdatePointBalance(ctx context.Context, userID string, delta int) error
}

// ARRepository はAR試着セッションの永続化を担う
type ARRepository interface {
	Create(ctx context.Context, session *domain.ARSession) error
	FindByID(ctx context.Context, id string) (*domain.ARSession, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.ARSession, error)
}

// MaterialRepository は素材マスタの永続化を担う
type MaterialRepository interface {
	List(ctx context.Context, category, textureType *string, limit int, cursor string) ([]*domain.Material, string, error)
	Create(ctx context.Context, m *domain.Material) error
	Update(ctx context.Context, m *domain.Material) error
	FindByID(ctx context.Context, id string) (*domain.Material, error)
}

// DesignRepository はデザインIPの永続化を担う
type DesignRepository interface {
	Create(ctx context.Context, design *domain.DesignIP) error
	FindByID(ctx context.Context, id string) (*domain.DesignIP, error)
	Update(ctx context.Context, design *domain.DesignIP) error

	// pgvector コサイン類似度検索（<=> 演算子）
	FindSimilar(ctx context.Context, vector []float32, forkThreshold, rejectThreshold float64) ([]SimilarDesignResult, error)

	ListFeed(ctx context.Context, filter DesignFeedFilter) ([]*domain.DesignIP, string, error)
	ListByCreator(ctx context.Context, creatorID string, cursor string, limit int) ([]*domain.DesignIP, string, error)

	CreateRoyaltyNode(ctx context.Context, node *domain.DesignRoyaltyNode) error
	ListRoyaltyNodes(ctx context.Context, designIPID string) ([]*domain.DesignRoyaltyNode, error)
	GetRoyaltyChain(ctx context.Context, designIPID string) ([]domain.DesignRoyaltyNode, error)
	IncrementUsageCount(ctx context.Context, designIPID string, royaltyYen int) error
	UpdateStatus(ctx context.Context, id string, status string) error
}

type SimilarDesignResult struct {
	Design *domain.DesignIP
	Score  float64
}

type DesignFeedFilter struct {
	GenderTag *string
	StyleTags []string
	Status    *string
	Sort      string
	Cursor    string
	Limit     int
}

// SalonRepository はサロン情報の永続化を担う
type SalonRepository interface {
	Create(ctx context.Context, salon *domain.Salon) error
	FindByID(ctx context.Context, id string) (*domain.Salon, error)
	FindByOwnerID(ctx context.Context, ownerID string) (*domain.Salon, error)
	Update(ctx context.Context, salon *domain.Salon) error
	List(ctx context.Context, prefecture *string, verificationStatus *string, skillTags []string, sort, cursor string, limit int) ([]*domain.Salon, string, error)
	// エリア・スキルタグにマッチするサロン一覧（入札通知用）
	FindMatchingSalons(ctx context.Context, prefecture string, requiredTags []string) ([]*domain.Salon, error)
}

// AuctionRepository は予約リクエスト・入札の永続化を担う
type AuctionRepository interface {
	CreateBookingRequest(ctx context.Context, req *domain.BookingRequest) error
	FindBookingRequestByID(ctx context.Context, id string) (*domain.BookingRequest, error)
	UpdateBookingRequestStatus(ctx context.Context, id string, status domain.BookingRequestStatus) error
	CountOpenRequestsByUser(ctx context.Context, userID string) (int, error)
	ListOpenRequestsForSalon(ctx context.Context, prefecture string, cursor string, limit int) ([]*domain.BookingRequest, string, error)

	CreateBid(ctx context.Context, bid *domain.Bid) error
	FindBidByID(ctx context.Context, id string) (*domain.Bid, error)
	FindBidBySalonAndRequest(ctx context.Context, salonID, requestID string) (*domain.Bid, error)
	UpdateBid(ctx context.Context, bid *domain.Bid) error
	ListBidsByRequestID(ctx context.Context, requestID string) ([]*domain.Bid, error)
}

// BookingRepository は確定予約の永続化を担う
type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	FindByID(ctx context.Context, id string) (*domain.Booking, error)
	ListByUserID(ctx context.Context, userID string, status *string) ([]*domain.Booking, error)
	ListBySalonID(ctx context.Context, salonID string, status *string) ([]*domain.Booking, error)
	UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error
}

// PaymentRepository は決済・ロイヤリティ分配の永続化を担う
type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	FindByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error
	CreateRoyaltyDistributions(ctx context.Context, paymentID string, dists []domain.RoyaltyDistributionItem) error
	UpdateRoyaltyStatus(ctx context.Context, id string, status domain.RoyaltyStatus) error
}

// ReviewRepository はレビューの永続化を担う
type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	FindByBookingID(ctx context.Context, bookingID string) (*domain.Review, error)
	ListBySalonID(ctx context.Context, salonID string, cursor string, limit int) ([]*domain.Review, string, error)
	UpdateAIScore(ctx context.Context, id string, score float64, status string) error
}

// NotificationRepository は通知の永続化を担う
type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	ListByUserID(ctx context.Context, userID string, isRead *bool, cursor string, limit int) ([]*domain.Notification, string, error)
	MarkRead(ctx context.Context, notificationID, userID string) error
}
