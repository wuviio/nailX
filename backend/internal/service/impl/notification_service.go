package impl

import (
	"context"

	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type notificationService struct {
	notifRepo repository.NotificationRepository
}

func NewNotificationService(notifRepo repository.NotificationRepository) service.NotificationService {
	return &notificationService{notifRepo: notifRepo}
}

func (s *notificationService) ListNotifications(ctx context.Context, userID string, isRead *bool, cursor string, limit int) ([]*domain.Notification, string, error) {
	return s.notifRepo.ListByUserID(ctx, userID, isRead, cursor, limit)
}

func (s *notificationService) MarkRead(ctx context.Context, userID, notificationID string) error {
	return s.notifRepo.MarkRead(ctx, notificationID, userID)
}

func (s *notificationService) RegisterFCMToken(ctx context.Context, userID, token, platform string) error {
	// TODO: FCMトークンをDBに保存（fcm_tokens テーブルが必要）
	return nil
}
