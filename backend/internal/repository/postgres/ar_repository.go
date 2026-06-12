package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type arRepository struct {
	db *pgxpool.Pool
}

func NewARRepository(db *pgxpool.Pool) repository.ARRepository {
	return &arRepository{db: db}
}

const arColumns = `id, user_id, design_ip_id, detected_nail_length_mm, has_existing_gel, detected_nail_shape,
	estimated_treatment_min, estimated_gel_amount_ml, hand_snapshot_url, created_at`

func (r *arRepository) Create(ctx context.Context, s *domain.ARSession) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO ar_sessions
			(id, user_id, design_ip_id, detected_nail_length_mm, has_existing_gel, detected_nail_shape,
			 estimated_treatment_min, estimated_gel_amount_ml, hand_snapshot_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		s.ID, s.UserID, s.DesignIPID, s.DetectedNailLengthMM, s.HasExistingGel, s.DetectedNailShape,
		s.EstimatedTreatmentMin, s.EstimatedGelAmountML, s.HandSnapshotURL,
	)
	return err
}

func (r *arRepository) FindByID(ctx context.Context, id string) (*domain.ARSession, error) {
	row := r.db.QueryRow(ctx, `SELECT `+arColumns+` FROM ar_sessions WHERE id=$1`, id)
	return scanARSession(row)
}

func (r *arRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.ARSession, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+arColumns+` FROM ar_sessions WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.ARSession
	for rows.Next() {
		s, err := scanARSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func scanARSession(row pgx.Row) (*domain.ARSession, error) {
	s := &domain.ARSession{}
	err := row.Scan(
		&s.ID, &s.UserID, &s.DesignIPID,
		&s.DetectedNailLengthMM, &s.HasExistingGel, &s.DetectedNailShape,
		&s.EstimatedTreatmentMin, &s.EstimatedGelAmountML, &s.HandSnapshotURL,
		&s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}
