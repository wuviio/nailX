package domain

import (
	"encoding/json"
	"time"
)

type BookingRequestStatus string

const (
	BookingRequestOpen      BookingRequestStatus = "open"
	BookingRequestBidding   BookingRequestStatus = "bidding"
	BookingRequestConfirmed BookingRequestStatus = "confirmed"
	BookingRequestCancelled BookingRequestStatus = "cancelled"
	BookingRequestExpired   BookingRequestStatus = "expired"
)

type BidStatus string

const (
	BidPending           BidStatus = "pending"
	BidAccepted          BidStatus = "accepted"
	BidRejected          BidStatus = "rejected"
	BidCancelledBySalon  BidStatus = "cancelled_by_salon"
	BidExpired           BidStatus = "expired"
)

type BookingStatus string

const (
	BookingConfirmed         BookingStatus = "confirmed"
	BookingCompleted         BookingStatus = "completed"
	BookingCancelledByUser   BookingStatus = "cancelled_by_user"
	BookingCancelledBySalon  BookingStatus = "cancelled_by_salon"
	BookingNoShow            BookingStatus = "no_show"
)

type NailDataSnapshot struct {
	LengthMM              *float64 `json:"length_mm,omitempty"`
	HasExistingGel        *bool    `json:"has_existing_gel,omitempty"`
	Shape                 *string  `json:"shape,omitempty"`
	EstimatedTreatmentMin *int     `json:"estimated_treatment_min,omitempty"`
	EstimatedGelAmountML  *float64 `json:"estimated_gel_amount_ml,omitempty"`
}

type BookingRequest struct {
	ID               string               `json:"id"`
	UserID           string               `json:"user_id"`
	DesignIPID       string               `json:"design_ip_id"`
	ARSessionID      *string              `json:"ar_session_id,omitempty"`
	NailDataSnapshot json.RawMessage      `json:"nail_data_snapshot"`
	BudgetMaxYen     int                  `json:"budget_max_yen"`
	DesiredDateFrom  time.Time            `json:"desired_date_from"`
	DesiredDateTo    time.Time            `json:"desired_date_to"`
	AreaPrefecture   string               `json:"area_prefecture"`
	AreaCity         *string              `json:"area_city,omitempty"`
	Status           BookingRequestStatus `json:"status"`
	ExpiresAt        time.Time            `json:"expires_at"`
	CreatedAt        time.Time            `json:"created_at"`
}

type Bid struct {
	ID                   string    `json:"id"`
	BookingRequestID     string    `json:"booking_request_id"`
	SalonID              string    `json:"salon_id"`
	PriceYen             int       `json:"price_yen"`
	IncludesRemoval      bool      `json:"includes_removal"`
	RemovalFeeYen        int       `json:"removal_fee_yen"`
	AvailableSlotAt      time.Time `json:"available_slot_at"`
	DynamicDiscountReason *string  `json:"dynamic_discount_reason,omitempty"`
	Message              *string   `json:"message,omitempty"`
	RebidCount           int       `json:"rebid_count"`
	Status               BidStatus `json:"status"`
	ExpiresAt            time.Time `json:"expires_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Booking struct {
	ID               string        `json:"id"`
	BookingRequestID string        `json:"booking_request_id"`
	BidID            string        `json:"bid_id"`
	UserID           string        `json:"user_id"`
	SalonID          string        `json:"salon_id"`
	ScheduledAt      time.Time     `json:"scheduled_at"`
	Status           BookingStatus `json:"status"`
	ConfirmedAt      time.Time     `json:"confirmed_at"`
	CompletedAt      *time.Time    `json:"completed_at,omitempty"`
	CancellationReason *string     `json:"cancellation_reason,omitempty"`
}

// SSE イベント型
type BidEvent struct {
	Type string `json:"type"` // "new_bid"
	Bid  *Bid   `json:"bid"`
}
