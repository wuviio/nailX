package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type AdminHandler struct {
	salonSvc  service.SalonService
	designSvc service.DesignService
}

func NewAdminHandler(salonSvc service.SalonService, designSvc service.DesignService) *AdminHandler {
	return &AdminHandler{salonSvc: salonSvc, designSvc: designSvc}
}

// GET /api/v1/admin/salons
func (h *AdminHandler) ListPendingSalons(c echo.Context) error {
	pending := "pending"
	salons, nextCur, err := h.salonSvc.ListSalons(c.Request().Context(), service.SalonFilter{
		VerificationStatus: &pending,
		Sort:               "created_at",
		Cursor:             c.QueryParam("cursor"),
		Limit:              20,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"salons": salons, "next_cursor": nextCur})
}

// PATCH /api/v1/admin/salons/:id/verify
func (h *AdminHandler) VerifySalon(c echo.Context) error {
	var req struct {
		Action string  `json:"action" validate:"required,oneof=approve reject"`
		Reason *string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := h.salonSvc.VerifySalon(c.Request().Context(), "admin", c.Param("id"), req.Action, req.Reason); err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "salon verification updated"})
}

// GET /api/v1/admin/designs/flagged
func (h *AdminHandler) ListFlaggedDesigns(c echo.Context) error {
	rejected := "rejected"
	designs, nextCur, err := h.designSvc.ListFeed(c.Request().Context(), service.DesignFeedFilter{
		Status: &rejected,
		Sort:   "created_at",
		Cursor: c.QueryParam("cursor"),
		Limit:  20,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"designs": designs, "next_cursor": nextCur})
}

// PATCH /api/v1/admin/designs/:id/moderate
func (h *AdminHandler) ModerateDesign(c echo.Context) error {
	var req struct {
		Action string  `json:"action" validate:"required,oneof=approve reject"`
		Reason *string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	// action を domain ステータスへ変換
	isPublic := req.Action == "approve"
	design, err := h.designSvc.UpdateDesign(c.Request().Context(), middleware.GetUserID(c), middleware.GetUserRole(c), c.Param("id"), service.UpdateDesignRequest{
		IsPublic: &isPublic,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"design": design})
}
