package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

type notificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, payload, is_read)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		n.ID, n.UserID, n.Type, n.Title, n.Body, n.Payload, n.IsRead,
	)
	return err
}

func (r *notificationRepository) ListByUserID(ctx context.Context, userID string, isRead *bool, cursor string, limit int) ([]*domain.Notification, string, error) {
	if limit == 0 {
		limit = 20
	}
	args := []any{userID}
	where := "WHERE user_id=$1"
	argIdx := 2
	if isRead != nil {
		where += " AND is_read=$" + itoa(argIdx)
		args = append(args, *isRead)
		argIdx++
	}
	if cursor != "" {
		ts, id, err := decodeCursor(cursor)
		if err == nil {
			where += " AND (created_at, id) < ($" + itoa(argIdx) + ", $" + itoa(argIdx+1) + ")"
			args = append(args, ts, id)
			argIdx += 2
		}
	}
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, type, title, body, payload, is_read, created_at
		FROM notifications `+where+`
		ORDER BY created_at DESC, id DESC LIMIT $`+itoa(len(args)),
		args...,
	)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var notifs []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Payload, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, "", err
		}
		notifs = append(notifs, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(notifs) > limit {
		notifs = notifs[:limit]
		nextCursor = encodeCursor(notifs[limit-1].CreatedAt, notifs[limit-1].ID)
	}
	return notifs, nextCursor, nil
}

func (r *notificationRepository) MarkRead(ctx context.Context, notificationID, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE notifications SET is_read=true WHERE id=$1 AND user_id=$2`,
		notificationID, userID)
	return err
}
