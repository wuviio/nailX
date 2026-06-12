package domain

import "time"

type UserRole string

const (
	RoleConsumer   UserRole = "consumer"
	RoleSalonOwner UserRole = "salon_owner"
	RoleAdmin      UserRole = "admin"
)

type User struct {
	ID            string    `json:"id"`
	FirebaseUID   string    `json:"-"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	Gender        *string   `json:"gender,omitempty"`
	Role          UserRole  `json:"role"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	Bio           *string   `json:"bio,omitempty"`
	LifestyleTags []string  `json:"lifestyle_tags"`
	PointBalance  int       `json:"point_balance"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type NailProfile struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	AvgNailLengthMM  *float64  `json:"avg_nail_length_mm,omitempty"`
	NailShape        *string   `json:"nail_shape,omitempty"`
	GelLiftTendency  *string   `json:"gel_lift_tendency,omitempty"`
	AllergyNotes     *string   `json:"allergy_notes,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ARSession struct {
	ID                     string    `json:"id"`
	UserID                 string    `json:"user_id"`
	DesignIPID             *string   `json:"design_ip_id,omitempty"`
	DetectedNailLengthMM   *float64  `json:"detected_nail_length_mm,omitempty"`
	HasExistingGel         *bool     `json:"has_existing_gel,omitempty"`
	DetectedNailShape      *string   `json:"detected_nail_shape,omitempty"`
	EstimatedTreatmentMin  *int      `json:"estimated_treatment_min,omitempty"`
	EstimatedGelAmountML   *float64  `json:"estimated_gel_amount_ml,omitempty"`
	HandSnapshotURL        *string   `json:"hand_snapshot_url,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
}
