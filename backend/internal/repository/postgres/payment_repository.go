package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type paymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) repository.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO payments
			(id, booking_id, total_amount_yen, platform_fee_yen, salon_payout_yen,
			 design_royalty_total_yen, stripe_payment_intent_id, stripe_charge_id,
			 payment_method, status, authorized_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		p.ID, p.BookingID, p.TotalAmountYen, p.PlatformFeeYen, p.SalonPayoutYen,
		p.DesignRoyaltyTotalYen, p.StripePaymentIntentID, p.StripeChargeID,
		p.PaymentMethod, p.Status, p.AuthorizedAt,
	)
	return err
}

func (r *paymentRepository) FindByBookingID(ctx context.Context, bookingID string) (*domain.Payment, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, booking_id, total_amount_yen, platform_fee_yen, salon_payout_yen,
		       design_royalty_total_yen, stripe_payment_intent_id, stripe_charge_id,
		       payment_method, status, authorized_at, captured_at, created_at
		FROM payments WHERE booking_id=$1`, bookingID)
	var p domain.Payment
	if err := row.Scan(
		&p.ID, &p.BookingID, &p.TotalAmountYen, &p.PlatformFeeYen, &p.SalonPayoutYen,
		&p.DesignRoyaltyTotalYen, &p.StripePaymentIntentID, &p.StripeChargeID,
		&p.PaymentMethod, &p.Status, &p.AuthorizedAt, &p.CapturedAt, &p.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE payments SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *paymentRepository) CreateRoyaltyDistributions(ctx context.Context, paymentID string, dists []domain.RoyaltyDistributionItem) error {
	batch := &pgx.Batch{}
	for _, d := range dists {
		batch.Queue(`
			INSERT INTO royalty_distributions
				(id, payment_id, design_ip_id, user_id, amount_yen, share_percent, depth_level, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,'pending')
			ON CONFLICT (payment_id, user_id) DO NOTHING`,
			uuid.NewString(), paymentID, d.DesignIPID, d.UserID, d.AmountYen, d.Percent, d.DepthLevel,
		)
	}
	results := r.db.SendBatch(ctx, batch)
	return results.Close()
}

func (r *paymentRepository) UpdateRoyaltyStatus(ctx context.Context, id string, status domain.RoyaltyStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE royalty_distributions SET status=$1 WHERE id=$2`, status, id)
	return err
}
