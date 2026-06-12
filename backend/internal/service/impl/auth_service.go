package impl

import (
	"context"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type authService struct {
	fbApp    *firebase.App
	userRepo repository.UserRepository
}

func NewAuthService(fbApp *firebase.App, userRepo repository.UserRepository) service.AuthService {
	return &authService{fbApp: fbApp, userRepo: userRepo}
}

func (s *authService) Register(ctx context.Context, firebaseToken, displayName string, gender *string) (*domain.User, error) {
	client, err := s.fbApp.Auth(ctx)
	if err != nil {
		return nil, err
	}
	token, err := client.VerifyIDToken(ctx, firebaseToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// 既存ユーザーチェック
	existing, err := s.userRepo.FindByFirebaseUID(ctx, token.UID)
	if err == nil {
		return existing, nil // 既に登録済み
	}
	if err != domain.ErrNotFound {
		return nil, err
	}

	email, _ := token.Claims["email"].(string)
	user := &domain.User{
		ID:          uuid.NewString(),
		FirebaseUID: token.UID,
		Email:       email,
		DisplayName: displayName,
		Gender:      gender,
		Role:        domain.RoleConsumer,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
