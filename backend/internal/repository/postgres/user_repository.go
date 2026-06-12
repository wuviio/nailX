package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, firebase_uid, email, display_name, avatar_url, role, point_balance)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.ID, u.FirebaseUID, u.Email, u.DisplayName, u.AvatarURL, u.Role, u.PointBalance,
	)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, firebase_uid, email, display_name, avatar_url, role, point_balance, created_at, updated_at
		FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func (r *userRepository) FindByFirebaseUID(ctx context.Context, uid string) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, firebase_uid, email, display_name, avatar_url, role, point_balance, created_at, updated_at
		FROM users WHERE firebase_uid = $1`, uid)
	return scanUser(row)
}

func (r *userRepository) Update(ctx context.Context, u *domain.User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET display_name=$1, avatar_url=$2 WHERE id=$3`,
		u.DisplayName, u.AvatarURL, u.ID,
	)
	return err
}

func (r *userRepository) UpsertNailProfile(ctx context.Context, p *domain.NailProfile) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO nail_profiles (id, user_id, nail_shape, avg_nail_length_mm, gel_lift_tendency, allergy_notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			nail_shape         = EXCLUDED.nail_shape,
			avg_nail_length_mm = EXCLUDED.avg_nail_length_mm,
			gel_lift_tendency  = EXCLUDED.gel_lift_tendency,
			allergy_notes      = EXCLUDED.allergy_notes,
			updated_at         = NOW()`,
		p.ID, p.UserID, p.NailShape, p.AvgNailLengthMM, p.GelLiftTendency, p.AllergyNotes,
	)
	return err
}

func (r *userRepository) GetNailProfile(ctx context.Context, userID string) (*domain.NailProfile, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, nail_shape, avg_nail_length_mm, gel_lift_tendency, allergy_notes, updated_at
		FROM nail_profiles WHERE user_id = $1`, userID)
	var p domain.NailProfile
	if err := row.Scan(&p.ID, &p.UserID, &p.NailShape, &p.AvgNailLengthMM, &p.GelLiftTendency, &p.AllergyNotes, &p.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *userRepository) UpdatePointBalance(ctx context.Context, userID string, delta int) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET point_balance = point_balance + $1 WHERE id = $2`, delta, userID)
	return err
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(
		&u.ID, &u.FirebaseUID, &u.Email, &u.DisplayName, &u.AvatarURL,
		&u.Role, &u.PointBalance, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
