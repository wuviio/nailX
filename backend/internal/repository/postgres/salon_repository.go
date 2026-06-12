package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type salonRepository struct {
	db *pgxpool.Pool
}

func NewSalonRepository(db *pgxpool.Pool) repository.SalonRepository {
	return &salonRepository{db: db}
}

const salonColumns = `id, owner_id, name, description, address, prefecture, city,
	lat, lng, phone, avg_reproduction_score, skill_badge_tags, portfolio_image_urls,
	verification_status, is_active, created_at, updated_at`

func (r *salonRepository) Create(ctx context.Context, s *domain.Salon) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO salons
			(id, owner_id, name, description, address, prefecture, city, lat, lng, phone,
			 skill_badge_tags, portfolio_image_urls, verification_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'pending')`,
		s.ID, s.OwnerID, s.Name, s.Description, s.Address, s.Prefecture, s.City,
		s.Lat, s.Lng, s.Phone, s.SkillBadgeTags, s.PortfolioImageURLs,
	)
	return err
}

func (r *salonRepository) FindByID(ctx context.Context, id string) (*domain.Salon, error) {
	row := r.db.QueryRow(ctx, `SELECT `+salonColumns+` FROM salons WHERE id=$1`, id)
	return scanSalon(row)
}

func (r *salonRepository) FindByOwnerID(ctx context.Context, ownerID string) (*domain.Salon, error) {
	row := r.db.QueryRow(ctx, `SELECT `+salonColumns+` FROM salons WHERE owner_id=$1`, ownerID)
	return scanSalon(row)
}

func (r *salonRepository) Update(ctx context.Context, s *domain.Salon) error {
	_, err := r.db.Exec(ctx, `
		UPDATE salons SET
			name=$1, description=$2, address=$3, city=$4,
			lat=$5, lng=$6, phone=$7, skill_badge_tags=$8, portfolio_image_urls=$9
		WHERE id=$10`,
		s.Name, s.Description, s.Address, s.City,
		s.Lat, s.Lng, s.Phone, s.SkillBadgeTags, s.PortfolioImageURLs, s.ID,
	)
	return err
}

func (r *salonRepository) List(ctx context.Context, prefecture *string, verificationStatus *string, skillTags []string, sort, cursor string, limit int) ([]*domain.Salon, string, error) {
	if limit == 0 {
		limit = 20
	}
	args := []any{}
	conditions := []string{}
	argIdx := 1

	if verificationStatus != nil {
		// 管理者ビュー: 指定ステータスのみ（is_active 条件なし）
		conditions = append(conditions, "verification_status=$"+itoa(argIdx))
		args = append(args, *verificationStatus)
		argIdx++
	} else {
		// 公開ビュー: 承認済み・アクティブのみ
		conditions = append(conditions, "is_active = true", "verification_status = 'approved'")
	}

	if prefecture != nil {
		conditions = append(conditions, "prefecture=$"+itoa(argIdx))
		args = append(args, *prefecture)
		argIdx++
	}
	if len(skillTags) > 0 {
		conditions = append(conditions, "skill_badge_tags && $"+itoa(argIdx))
		args = append(args, skillTags)
		argIdx++
	}
	if cursor != "" {
		ts, id, err := decodeCursor(cursor)
		if err == nil {
			conditions = append(conditions, "(created_at, id) < ($"+itoa(argIdx)+", $"+itoa(argIdx+1)+")")
			args = append(args, ts, id)
			argIdx += 2
		}
	}

	where := "WHERE " + strings.Join(conditions, " AND ")
	args = append(args, limit+1)

	orderBy := "created_at DESC, id DESC"
	if sort == "score" {
		orderBy = "avg_reproduction_score DESC, created_at DESC, id DESC"
	}

	rows, err := r.db.Query(ctx,
		`SELECT `+salonColumns+` FROM salons `+where+` ORDER BY `+orderBy+` LIMIT $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var salons []*domain.Salon
	for rows.Next() {
		s, err := scanSalon(rows)
		if err != nil {
			return nil, "", err
		}
		salons = append(salons, s)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(salons) > limit {
		salons = salons[:limit]
		nextCursor = encodeCursor(salons[limit-1].CreatedAt, salons[limit-1].ID)
	}
	return salons, nextCursor, nil
}

func (r *salonRepository) FindMatchingSalons(ctx context.Context, prefecture string, requiredTags []string) ([]*domain.Salon, error) {
	args := []any{prefecture}
	where := "WHERE prefecture=$1 AND is_active=true AND verification_status='approved'"
	if len(requiredTags) > 0 {
		where += " AND skill_badge_tags && $2"
		args = append(args, requiredTags)
	}
	rows, err := r.db.Query(ctx, `SELECT `+salonColumns+` FROM salons `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var salons []*domain.Salon
	for rows.Next() {
		s, err := scanSalon(rows)
		if err != nil {
			return nil, err
		}
		salons = append(salons, s)
	}
	return salons, rows.Err()
}

func scanSalon(row pgx.Row) (*domain.Salon, error) {
	var s domain.Salon
	if err := row.Scan(
		&s.ID, &s.OwnerID, &s.Name, &s.Description, &s.Address, &s.Prefecture, &s.City,
		&s.Lat, &s.Lng, &s.Phone, &s.AvgReproductionScore, &s.SkillBadgeTags, &s.PortfolioImageURLs,
		&s.VerificationStatus, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}
