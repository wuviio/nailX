package domain

import "time"

type PaymentStatus string

const (
	PaymentAuthorized PaymentStatus = "authorized"
	PaymentCaptured   PaymentStatus = "captured"
	PaymentRefunded   PaymentStatus = "refunded"
	PaymentFailed     PaymentStatus = "failed"
)

type RoyaltyStatus string

const (
	RoyaltyPending  RoyaltyStatus = "pending"
	RoyaltyCredited RoyaltyStatus = "credited"
	RoyaltyPaid     RoyaltyStatus = "paid"
	RoyaltyCancelled RoyaltyStatus = "cancelled"
)

const (
	PlatformFeeRate  = 0.15 // 15%
	IPUsageFeeRate   = 0.05 // 5%
)

type Payment struct {
	ID                       string        `json:"id"`
	BookingID                string        `json:"booking_id"`
	TotalAmountYen           int           `json:"total_amount_yen"`
	PlatformFeeYen           int           `json:"platform_fee_yen"`
	SalonPayoutYen           int           `json:"salon_payout_yen"`
	DesignRoyaltyTotalYen    int           `json:"design_royalty_total_yen"`
	StripePaymentIntentID    *string       `json:"stripe_payment_intent_id,omitempty"`
	StripeChargeID           *string       `json:"stripe_charge_id,omitempty"`
	PaymentMethod            *string       `json:"payment_method,omitempty"`
	Status                   PaymentStatus `json:"status"`
	AuthorizedAt             time.Time     `json:"authorized_at"`
	CapturedAt               *time.Time    `json:"captured_at,omitempty"`
	CreatedAt                time.Time     `json:"created_at"`
}

type RoyaltyDistribution struct {
	ID           string        `json:"id"`
	PaymentID    string        `json:"payment_id"`
	DesignIPID   string        `json:"design_ip_id"`
	UserID       string        `json:"user_id"`
	AmountYen    int           `json:"amount_yen"`
	SharePercent float64       `json:"share_percent"`
	DepthLevel   int           `json:"depth_level"`
	Status       RoyaltyStatus `json:"status"`
	CreditedAt   *time.Time    `json:"credited_at,omitempty"`
	PaidAt       *time.Time    `json:"paid_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// CalcPaymentBreakdown は施術代から各費用を計算する（切り捨て）
func CalcPaymentBreakdown(totalYen int) (platformFee, ipFee, salonPayout int) {
	platformFee = int(float64(totalYen) * PlatformFeeRate)
	ipFee = int(float64(totalYen) * IPUsageFeeRate)
	salonPayout = totalYen - platformFee - ipFee
	return
}

type Salon struct {
	ID                   string    `json:"id"`
	OwnerID              string    `json:"owner_id"`
	Name                 string    `json:"name"`
	Description          *string   `json:"description,omitempty"`
	Address              string    `json:"address"`
	Prefecture           string    `json:"prefecture"`
	City                 *string   `json:"city,omitempty"`
	Lat                  *float64  `json:"lat,omitempty"`
	Lng                  *float64  `json:"lng,omitempty"`
	Phone                *string   `json:"phone,omitempty"`
	AvgReproductionScore float64   `json:"avg_reproduction_score"`
	SkillBadgeTags       []string  `json:"skill_badge_tags"`
	PortfolioImageURLs   []string  `json:"portfolio_image_urls"`
	VerificationStatus   string    `json:"verification_status"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Review struct {
	ID                  string     `json:"id"`
	BookingID           string     `json:"booking_id"`
	UserID              string     `json:"user_id"`
	SalonID             string     `json:"salon_id"`
	DesignIPID          string     `json:"design_ip_id"`
	ReproductionScore   int        `json:"reproduction_score"`
	OverallScore        int        `json:"overall_score"`
	Comment             *string    `json:"comment,omitempty"`
	BeforePhotoURL      *string    `json:"before_photo_url,omitempty"`
	AfterPhotoURL       *string    `json:"after_photo_url,omitempty"`
	AIReproductionScore *float64   `json:"ai_reproduction_score,omitempty"`
	AIAnalysisStatus    string     `json:"ai_analysis_status"`
	CreatedAt           time.Time  `json:"created_at"`
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      *string   `json:"body,omitempty"`
	Payload   any       `json:"payload"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
