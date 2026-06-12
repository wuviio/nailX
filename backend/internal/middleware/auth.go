package middleware

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/domain"
)

type contextKey string

const (
	contextKeyFirebaseUID contextKey = "firebase_uid"
	contextKeyUserID      contextKey = "user_id"
	contextKeyUserRole    contextKey = "user_role"
)

type AuthMiddleware struct {
	firebaseAuth *auth.Client
	userRepo     UserRepository
}

// UserRepository は認証ミドルウェアが必要とする最小限のリポジトリインターフェース
type UserRepository interface {
	FindByFirebaseUID(ctx context.Context, uid string) (*domain.User, error)
}

func NewAuthMiddleware(app *firebase.App, userRepo UserRepository) (*AuthMiddleware, error) {
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return &AuthMiddleware{firebaseAuth: client, userRepo: userRepo}, nil
}

// RequireAuth は Firebase ID Token を検証し、DBユーザー情報をコンテキストに注入する
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := extractBearerToken(c.Request())
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{
				"error": map[string]string{"code": "UNAUTHORIZED", "message": "missing or invalid authorization header"},
			})
		}

		firebaseToken, err := m.firebaseAuth.VerifyIDToken(c.Request().Context(), token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{
				"error": map[string]string{"code": "UNAUTHORIZED", "message": "invalid firebase token"},
			})
		}

		user, err := m.userRepo.FindByFirebaseUID(c.Request().Context(), firebaseToken.UID)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]any{
				"error": map[string]string{"code": "UNAUTHORIZED", "message": "user not found"},
			})
		}

		c.Set(string(contextKeyFirebaseUID), firebaseToken.UID)
		c.Set(string(contextKeyUserID), user.ID)
		c.Set(string(contextKeyUserRole), string(user.Role))
		return next(c)
	}
}

// RequireRole は指定ロール以上の権限を要求する
func RequireRole(role domain.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := domain.UserRole(c.Get(string(contextKeyUserRole)).(string))
			if !hasRole(userRole, role) {
				return echo.NewHTTPError(http.StatusForbidden, map[string]any{
					"error": map[string]string{"code": "FORBIDDEN", "message": "insufficient permissions"},
				})
			}
			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	v, _ := c.Get(string(contextKeyUserID)).(string)
	return v
}

func GetUserRole(c echo.Context) domain.UserRole {
	v, _ := c.Get(string(contextKeyUserRole)).(string)
	return domain.UserRole(v)
}

func extractBearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return "", domain.ErrUnauthorized
	}
	return strings.TrimPrefix(h, "Bearer "), nil
}

func hasRole(userRole, required domain.UserRole) bool {
	order := map[domain.UserRole]int{
		domain.RoleConsumer:   1,
		domain.RoleSalonOwner: 2,
		domain.RoleAdmin:      3,
	}
	return order[userRole] >= order[required]
}
