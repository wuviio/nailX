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

type designRepository struct {
	db *pgxpool.Pool
}

func NewDesignRepository(db *pgxpool.Pool) repository.DesignRepository {
	return &designRepository{db: db}
}

func (r *designRepository) Create(ctx context.Context, d *domain.DesignIP) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO design_ips
			(id, creator_id, parent_ip_id, title, description, preview_image_url, design_data,
			 fork_depth, status, is_public, gender_tag, style_tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		d.ID, d.CreatorID, d.ParentIPID, d.Title, d.Description, d.PreviewImageURL, d.DesignData,
		d.ForkDepth, d.Status, d.IsPublic, d.GenderTag, d.StyleTags,
	)
	return err
}

func (r *designRepository) FindByID(ctx context.Context, id string) (*domain.DesignIP, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, creator_id, parent_ip_id, title, description, preview_image_url, design_data,
		       fork_depth, status, is_public, gender_tag, style_tags,
		       usage_count, total_royalty_yen, created_at, updated_at
		FROM design_ips WHERE id = $1`, id)
	return scanDesign(row)
}

func (r *designRepository) Update(ctx context.Context, d *domain.DesignIP) error {
	_, err := r.db.Exec(ctx, `
		UPDATE design_ips SET
			title=$1, description=$2, preview_image_url=$3,
			status=$4, is_public=$5, gender_tag=$6, style_tags=$7
		WHERE id=$8`,
		d.Title, d.Description, d.PreviewImageURL,
		d.Status, d.IsPublic, d.GenderTag, d.StyleTags, d.ID,
	)
	return err
}

func (r *designRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE design_ips SET status=$1 WHERE id=$2`, status, id)
	return err
}

// FindSimilar はコサイン類似度（<=> 演算子）でベクトル近傍を検索し、
// forkThreshold 以上のものを返す（rejectThreshold 以上は先頭に来る）
func (r *designRepository) FindSimilar(ctx context.Context, vector []float32, forkThreshold, rejectThreshold float64) ([]repository.SimilarDesignResult, error) {
	// pgvector の <=> はコサイン距離（0=同一, 2=逆）なので類似度 = 1 - 距離
	rows, err := r.db.Query(ctx, `
		SELECT id, creator_id, parent_ip_id, title, description, preview_image_url, design_data,
		       fork_depth, status, is_public, gender_tag, style_tags,
		       usage_count, total_royalty_yen, created_at, updated_at,
		       (1 - (similarity_vector <=> $1::vector)) AS cosine_score
		FROM design_ips
		WHERE (1 - (similarity_vector <=> $1::vector)) >= $2
		  AND status = 'active'
		ORDER BY cosine_score DESC
		LIMIT 10`,
		vectorToString(vector), forkThreshold,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []repository.SimilarDesignResult
	for rows.Next() {
		var d domain.DesignIP
		var score float64
		if err := rows.Scan(
			&d.ID, &d.CreatorID, &d.ParentIPID, &d.Title, &d.Description, &d.PreviewImageURL, &d.DesignData,
			&d.ForkDepth, &d.Status, &d.IsPublic, &d.GenderTag, &d.StyleTags,
			&d.UsageCount, &d.TotalRoyaltyYen, &d.CreatedAt, &d.UpdatedAt,
			&score,
		); err != nil {
			return nil, err
		}
		results = append(results, repository.SimilarDesignResult{Design: &d, Score: score})
	}
	return results, rows.Err()
}

func (r *designRepository) ListFeed(ctx context.Context, filter repository.DesignFeedFilter) ([]*domain.DesignIP, string, error) {
	// カーソルページネーション: base64url(created_at_unix_ms + ":" + id)
	args := []any{}
	argIdx := 1

	// 管理者専用ステータス（rejected / pending）指定時は is_public フィルターを外す
	adminOnlyStatus := map[string]bool{"rejected": true, "pending": true}
	isAdminQuery := filter.Status != nil && adminOnlyStatus[*filter.Status]

	conditions := []string{}
	if !isAdminQuery {
		conditions = append(conditions, "d.is_public = true")
	}

	if filter.Status != nil {
		conditions = append(conditions, "d.status = $"+itoa(argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.GenderTag != nil {
		conditions = append(conditions, "d.gender_tag = $"+itoa(argIdx))
		args = append(args, *filter.GenderTag)
		argIdx++
	}

	if filter.Cursor != "" {
		ts, id, err := decodeCursor(filter.Cursor)
		if err == nil {
			conditions = append(conditions, "(d.created_at, d.id) < ($"+itoa(argIdx)+", $"+itoa(argIdx+1)+")")
			args = append(args, ts, id)
			argIdx += 2
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := filter.Limit
	if limit == 0 {
		limit = 20
	}
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, `
		SELECT id, creator_id, parent_ip_id, title, description, preview_image_url, design_data,
		       fork_depth, status, is_public, gender_tag, style_tags,
		       usage_count, total_royalty_yen, created_at, updated_at
		FROM design_ips d `+where+`
		ORDER BY d.created_at DESC, d.id DESC
		LIMIT $`+itoa(argIdx),
		args...,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	designs, err := scanDesigns(rows)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(designs) > limit {
		designs = designs[:limit]
		last := designs[limit-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}
	return designs, nextCursor, nil
}

func (r *designRepository) ListByCreator(ctx context.Context, creatorID string, cursor string, limit int) ([]*domain.DesignIP, string, error) {
	if limit == 0 {
		limit = 20
	}
	args := []any{creatorID}
	where := "WHERE creator_id = $1"
	if cursor != "" {
		ts, id, err := decodeCursor(cursor)
		if err == nil {
			where += " AND (created_at, id) < ($2, $3)"
			args = append(args, ts, id)
		}
	}
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, `
		SELECT id, creator_id, parent_ip_id, title, description, preview_image_url, design_data,
		       fork_depth, status, is_public, gender_tag, style_tags,
		       usage_count, total_royalty_yen, created_at, updated_at
		FROM design_ips `+where+`
		ORDER BY created_at DESC, id DESC LIMIT $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	designs, err := scanDesigns(rows)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(designs) > limit {
		designs = designs[:limit]
		nextCursor = encodeCursor(designs[limit-1].CreatedAt, designs[limit-1].ID)
	}
	return designs, nextCursor, nil
}

func (r *designRepository) CreateRoyaltyNode(ctx context.Context, node *domain.DesignRoyaltyNode) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO design_royalty_nodes (id, design_ip_id, user_id, share_percent, depth_level)
		VALUES ($1, $2, $3, $4, $5)`,
		node.ID, node.DesignIPID, node.UserID, node.SharePercent, node.DepthLevel,
	)
	return err
}

func (r *designRepository) ListRoyaltyNodes(ctx context.Context, designIPID string) ([]*domain.DesignRoyaltyNode, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, design_ip_id, user_id, share_percent, depth_level
		FROM design_royalty_nodes WHERE design_ip_id = $1 ORDER BY depth_level`, designIPID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nodes []*domain.DesignRoyaltyNode
	for rows.Next() {
		var n domain.DesignRoyaltyNode
		if err := rows.Scan(&n.ID, &n.DesignIPID, &n.UserID, &n.SharePercent, &n.DepthLevel); err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}
	return nodes, rows.Err()
}

func (r *designRepository) GetRoyaltyChain(ctx context.Context, designIPID string) ([]domain.DesignRoyaltyNode, error) {
	nodes, err := r.ListRoyaltyNodes(ctx, designIPID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.DesignRoyaltyNode, len(nodes))
	for i, n := range nodes {
		result[i] = *n
	}
	return result, nil
}

func (r *designRepository) IncrementUsageCount(ctx context.Context, designIPID string, royaltyYen int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE design_ips SET
			usage_count = usage_count + 1,
			total_royalty_yen = total_royalty_yen + $1
		WHERE id = $2`, royaltyYen, designIPID)
	return err
}

func scanDesign(row pgx.Row) (*domain.DesignIP, error) {
	var d domain.DesignIP
	if err := row.Scan(
		&d.ID, &d.CreatorID, &d.ParentIPID, &d.Title, &d.Description, &d.PreviewImageURL, &d.DesignData,
		&d.ForkDepth, &d.Status, &d.IsPublic, &d.GenderTag, &d.StyleTags,
		&d.UsageCount, &d.TotalRoyaltyYen, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func scanDesigns(rows pgx.Rows) ([]*domain.DesignIP, error) {
	var designs []*domain.DesignIP
	for rows.Next() {
		var d domain.DesignIP
		if err := rows.Scan(
			&d.ID, &d.CreatorID, &d.ParentIPID, &d.Title, &d.Description, &d.PreviewImageURL, &d.DesignData,
			&d.ForkDepth, &d.Status, &d.IsPublic, &d.GenderTag, &d.StyleTags,
			&d.UsageCount, &d.TotalRoyaltyYen, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		designs = append(designs, &d)
	}
	return designs, rows.Err()
}
