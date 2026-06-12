package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// GET /api/v1/users/me
func (h *UserHandler) GetMe(c echo.Context) error {
	user, profile, err := h.svc.GetMe(c.Request().Context(), middleware.GetUserID(c))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"user":         user,
		"nail_profile": profile,
	})
}

// PATCH /api/v1/users/me
func (h *UserHandler) UpdateMe(c echo.Context) error {
	var req struct {
		DisplayName   *string  `json:"display_name"`
		AvatarURL     *string  `json:"avatar_url"`
		Bio           *string  `json:"bio"`
		LifestyleTags []string `json:"lifestyle_tags"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	user, err := h.svc.UpdateMe(c.Request().Context(), middleware.GetUserID(c), service.UpdateUserRequest{
		DisplayName:   req.DisplayName,
		AvatarURL:     req.AvatarURL,
		Bio:           req.Bio,
		LifestyleTags: req.LifestyleTags,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"user": user})
}

// GET /api/v1/users/:id
func (h *UserHandler) GetPublicProfile(c echo.Context) error {
	user, err := h.svc.GetPublicProfile(c.Request().Context(), c.Param("id"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"user": user})
}

// PUT /api/v1/users/me/nail-profile
func (h *UserHandler) UpsertNailProfile(c echo.Context) error {
	var req struct {
		AvgNailLengthMM *float64 `json:"avg_nail_length_mm"`
		NailShape       *string  `json:"nail_shape"`
		GelLiftTendency *string  `json:"gel_lift_tendency"`
		AllergyNotes    *string  `json:"allergy_notes"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	profile, err := h.svc.UpsertNailProfile(c.Request().Context(), middleware.GetUserID(c), service.UpsertNailProfileRequest{
		AvgNailLengthMM: req.AvgNailLengthMM,
		NailShape:       req.NailShape,
		GelLiftTendency: req.GelLiftTendency,
		AllergyNotes:    req.AllergyNotes,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"nail_profile": profile})
}
