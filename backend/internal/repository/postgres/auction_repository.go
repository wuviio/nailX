package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type auctionRepository struct {
	db *pgxpool.Pool
}

func NewAuctionRepository(db *pgxpool.Pool) repository.AuctionRepository {
	return &auctionRepository{db: db}
}

func (r *auctionRepository) CreateBookingRequest(ctx context.Context, req *domain.BookingRequest) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO booking_requests
			(id, user_id, design_ip_id, ar_session_id, nail_data_snapshot,
			 desired_date_from, desired_date_to, area_prefecture, area_city,
			 budget_max_yen, status, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		req.ID, req.UserID, req.DesignIPID, req.ARSessionID, req.NailDataSnapshot,
		req.DesiredDateFrom, req.DesiredDateTo, req.AreaPrefecture, req.AreaCity,
		req.BudgetMaxYen, req.Status, req.ExpiresAt,
	)
	return err
}

func (r *auctionRepository) FindBookingRequestByID(ctx context.Context, id string) (*domain.BookingRequest, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, design_ip_id, ar_session_id, nail_data_snapshot,
		       desired_date_from, desired_date_to, area_prefecture, area_city,
		       budget_max_yen, status, expires_at, created_at
		FROM booking_requests WHERE id = $1`, id)
	var req domain.BookingRequest
	if err := row.Scan(
		&req.ID, &req.UserID, &req.DesignIPID, &req.ARSessionID, &req.NailDataSnapshot,
		&req.DesiredDateFrom, &req.DesiredDateTo, &req.AreaPrefecture, &req.AreaCity,
		&req.BudgetMaxYen, &req.Status, &req.ExpiresAt, &req.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &req, nil
}

func (r *auctionRepository) UpdateBookingRequestStatus(ctx context.Context, id string, status domain.BookingRequestStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE booking_requests SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *auctionRepository) CountOpenRequestsByUser(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM booking_requests
		WHERE user_id=$1 AND status='open'`, userID).Scan(&count)
	return count, err
}

func (r *auctionRepository) ListOpenRequestsForSalon(ctx context.Context, prefecture string, cursor string, limit int) ([]*domain.BookingRequest, string, error) {
	if limit == 0 {
		limit = 20
	}
	args := []any{prefecture}
	where := "WHERE area_prefecture=$1 AND status='open'"
	if cursor != "" {
		ts, id, err := decodeCursor(cursor)
		if err == nil {
			where += " AND (created_at, id) < ($2, $3)"
			args = append(args, ts, id)
		}
	}
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, design_ip_id, ar_session_id, nail_data_snapshot,
		       desired_date_from, desired_date_to, area_prefecture, area_city,
		       budget_max_yen, status, expires_at, created_at
		FROM booking_requests `+where+`
		ORDER BY created_at DESC, id DESC LIMIT $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var reqs []*domain.BookingRequest
	for rows.Next() {
		var req domain.BookingRequest
		if err := rows.Scan(
			&req.ID, &req.UserID, &req.DesignIPID, &req.ARSessionID, &req.NailDataSnapshot,
			&req.DesiredDateFrom, &req.DesiredDateTo, &req.AreaPrefecture, &req.AreaCity,
			&req.BudgetMaxYen, &req.Status, &req.ExpiresAt, &req.CreatedAt,
		); err != nil {
			return nil, "", err
		}
		reqs = append(reqs, &req)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(reqs) > limit {
		reqs = reqs[:limit]
		nextCursor = encodeCursor(reqs[limit-1].CreatedAt, reqs[limit-1].ID)
	}
	return reqs, nextCursor, nil
}

func (r *auctionRepository) CreateBid(ctx context.Context, bid *domain.Bid) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO bids
			(id, booking_request_id, salon_id, price_yen, includes_removal, removal_fee_yen,
			 available_slot_at, dynamic_discount_reason, message, status, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		bid.ID, bid.BookingRequestID, bid.SalonID, bid.PriceYen,
		bid.IncludesRemoval, bid.RemovalFeeYen, bid.AvailableSlotAt,
		bid.DynamicDiscountReason, bid.Message, bid.Status, bid.ExpiresAt,
	)
	return err
}

const bidColumns = `id, booking_request_id, salon_id, price_yen, includes_removal, removal_fee_yen,
	available_slot_at, dynamic_discount_reason, message, status, rebid_count, expires_at, created_at, updated_at`

func (r *auctionRepository) FindBidByID(ctx context.Context, id string) (*domain.Bid, error) {
	row := r.db.QueryRow(ctx, `SELECT `+bidColumns+` FROM bids WHERE id=$1`, id)
	return scanBid(row)
}

func (r *auctionRepository) FindBidBySalonAndRequest(ctx context.Context, salonID, requestID string) (*domain.Bid, error) {
	row := r.db.QueryRow(ctx, `SELECT `+bidColumns+` FROM bids WHERE salon_id=$1 AND booking_request_id=$2`, salonID, requestID)
	return scanBid(row)
}

func (r *auctionRepository) UpdateBid(ctx context.Context, bid *domain.Bid) error {
	_, err := r.db.Exec(ctx, `
		UPDATE bids SET
			price_yen=$1, includes_removal=$2, removal_fee_yen=$3,
			available_slot_at=$4, dynamic_discount_reason=$5, message=$6,
			status=$7, rebid_count=rebid_count+1
		WHERE id=$8`,
		bid.PriceYen, bid.IncludesRemoval, bid.RemovalFeeYen,
		bid.AvailableSlotAt, bid.DynamicDiscountReason, bid.Message,
		bid.Status, bid.ID,
	)
	return err
}

func (r *auctionRepository) ListBidsByRequestID(ctx context.Context, requestID string) ([]*domain.Bid, error) {
	rows, err := r.db.Query(ctx, `SELECT `+bidColumns+` FROM bids WHERE booking_request_id=$1 ORDER BY price_yen ASC`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bids []*domain.Bid
	for rows.Next() {
		b, err := scanBid(rows)
		if err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}
	return bids, rows.Err()
}

func scanBid(row pgx.Row) (*domain.Bid, error) {
	var b domain.Bid
	if err := row.Scan(
		&b.ID, &b.BookingRequestID, &b.SalonID, &b.PriceYen, &b.IncludesRemoval, &b.RemovalFeeYen,
		&b.AvailableSlotAt, &b.DynamicDiscountReason, &b.Message,
		&b.Status, &b.RebidCount, &b.ExpiresAt, &b.CreatedAt, &b.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}
