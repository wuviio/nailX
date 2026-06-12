package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/service"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req struct {
		FirebaseToken string  `json:"firebase_token" validate:"required"`
		DisplayName   string  `json:"display_name" validate:"required,max=100"`
		Gender        *string `json:"gender"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}

	user, err := h.svc.Register(c.Request().Context(), req.FirebaseToken, req.DisplayName, req.Gender)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"user": user})
}
