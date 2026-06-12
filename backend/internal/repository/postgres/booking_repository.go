package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type bookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) repository.BookingRepository {
	return &bookingRepository{db: db}
}

const bookingColumns = `id, booking_request_id, bid_id, user_id, salon_id, scheduled_at, status, confirmed_at, completed_at, cancellation_reason`

func (r *bookingRepository) Create(ctx context.Context, b *domain.Booking) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO bookings (id, booking_request_id, bid_id, user_id, salon_id, scheduled_at, status, confirmed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		b.ID, b.BookingRequestID, b.BidID, b.UserID, b.SalonID, b.ScheduledAt, b.Status, b.ConfirmedAt,
	)
	return err
}

func (r *bookingRepository) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	row := r.db.QueryRow(ctx, `SELECT `+bookingColumns+` FROM bookings WHERE id=$1`, id)
	return scanBooking(row)
}

func (r *bookingRepository) ListByUserID(ctx context.Context, userID string, status *string) ([]*domain.Booking, error) {
	args := []any{userID}
	where := "WHERE user_id=$1"
	if status != nil {
		where += " AND status=$2"
		args = append(args, *status)
	}
	rows, err := r.db.Query(ctx, `SELECT `+bookingColumns+` FROM bookings `+where+` ORDER BY scheduled_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBookings(rows)
}

func (r *bookingRepository) ListBySalonID(ctx context.Context, salonID string, status *string) ([]*domain.Booking, error) {
	args := []any{salonID}
	where := "WHERE salon_id=$1"
	if status != nil {
		where += " AND status=$2"
		args = append(args, *status)
	}
	rows, err := r.db.Query(ctx, `SELECT `+bookingColumns+` FROM bookings `+where+` ORDER BY scheduled_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBookings(rows)
}

func (r *bookingRepository) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE bookings SET status=$1 WHERE id=$2`, status, id)
	return err
}

func scanBooking(row pgx.Row) (*domain.Booking, error) {
	var b domain.Booking
	if err := row.Scan(
		&b.ID, &b.BookingRequestID, &b.BidID, &b.UserID, &b.SalonID,
		&b.ScheduledAt, &b.Status, &b.ConfirmedAt, &b.CompletedAt, &b.CancellationReason,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func scanBookings(rows pgx.Rows) ([]*domain.Booking, error) {
	var bookings []*domain.Booking
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}
