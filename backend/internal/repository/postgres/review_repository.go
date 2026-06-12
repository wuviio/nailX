package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type reviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

const reviewColumns = `id, booking_id, user_id, salon_id, design_ip_id,
	reproduction_score, overall_score, comment, before_photo_url, after_photo_url,
	ai_reproduction_score, ai_analysis_status, created_at`

func (r *reviewRepository) Create(ctx context.Context, rev *domain.Review) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO reviews
			(id, booking_id, user_id, salon_id, design_ip_id,
			 reproduction_score, overall_score, comment, before_photo_url, after_photo_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		rev.ID, rev.BookingID, rev.UserID, rev.SalonID, rev.DesignIPID,
		rev.ReproductionScore, rev.OverallScore, rev.Comment, rev.BeforePhotoURL, rev.AfterPhotoURL,
	)
	return err
}

func (r *reviewRepository) FindByBookingID(ctx context.Context, bookingID string) (*domain.Review, error) {
	row := r.db.QueryRow(ctx, `SELECT `+reviewColumns+` FROM reviews WHERE booking_id=$1`, bookingID)
	return scanReview(row)
}

func (r *reviewRepository) ListBySalonID(ctx context.Context, salonID string, cursor string, limit int) ([]*domain.Review, string, error) {
	if limit == 0 {
		limit = 20
	}
	args := []any{salonID}
	where := "WHERE salon_id=$1"
	if cursor != "" {
		ts, id, err := decodeCursor(cursor)
		if err == nil {
			where += " AND (created_at, id) < ($2, $3)"
			args = append(args, ts, id)
		}
	}
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, `
		SELECT `+reviewColumns+` FROM reviews `+where+`
		ORDER BY created_at DESC, id DESC LIMIT $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		rev, err := scanReview(rows)
		if err != nil {
			return nil, "", err
		}
		reviews = append(reviews, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(reviews) > limit {
		reviews = reviews[:limit]
		nextCursor = encodeCursor(reviews[limit-1].CreatedAt, reviews[limit-1].ID)
	}
	return reviews, nextCursor, nil
}

func (r *reviewRepository) UpdateAIScore(ctx context.Context, id string, score float64, status string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE reviews SET ai_reproduction_score=$1, ai_analysis_status=$2 WHERE id=$3`,
		score, status, id)
	return err
}

func scanReview(row pgx.Row) (*domain.Review, error) {
	var rev domain.Review
	if err := row.Scan(
		&rev.ID, &rev.BookingID, &rev.UserID, &rev.SalonID, &rev.DesignIPID,
		&rev.ReproductionScore, &rev.OverallScore, &rev.Comment, &rev.BeforePhotoURL, &rev.AfterPhotoURL,
		&rev.AIReproductionScore, &rev.AIAnalysisStatus, &rev.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &rev, nil
}
