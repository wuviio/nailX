package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type SalonHandler struct {
	svc service.SalonService
}

func NewSalonHandler(svc service.SalonService) *SalonHandler {
	return &SalonHandler{svc: svc}
}

// POST /api/v1/salons
func (h *SalonHandler) Register(c echo.Context) error {
	var req struct {
		Name          string         `json:"name" validate:"required,max=200"`
		Description   *string        `json:"description"`
		Address       string         `json:"address" validate:"required"`
		Prefecture    string         `json:"prefecture" validate:"required"`
		City          *string        `json:"city"`
		Lat           *float64       `json:"lat"`
		Lng           *float64       `json:"lng"`
		Phone         *string        `json:"phone"`
		BusinessHours map[string]any `json:"business_hours"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	hoursBytes, _ := json.Marshal(req.BusinessHours)
	salon, err := h.svc.RegisterSalon(c.Request().Context(), middleware.GetUserID(c), service.RegisterSalonRequest{
		Name:          req.Name,
		Description:   req.Description,
		Address:       req.Address,
		Prefecture:    req.Prefecture,
		City:          req.City,
		Lat:           req.Lat,
		Lng:           req.Lng,
		Phone:         req.Phone,
		BusinessHours: hoursBytes,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusAccepted, map[string]any{
		"salon_id":            salon.ID,
		"verification_status": salon.VerificationStatus,
	})
}

// GET /api/v1/salons
func (h *SalonHandler) List(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 20
	}
	salons, nextCur, err := h.svc.ListSalons(c.Request().Context(), service.SalonFilter{
		Prefecture: strPtr(c.QueryParam("prefecture")),
		Sort:       c.QueryParam("sort"),
		Cursor:     c.QueryParam("cursor"),
		Limit:      limit,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"salons": salons, "next_cursor": nextCur})
}

// GET /api/v1/salons/:id
func (h *SalonHandler) Get(c echo.Context) error {
	salon, err := h.svc.GetSalon(c.Request().Context(), c.Param("id"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"salon": salon})
}

// PATCH /api/v1/salons/:id
func (h *SalonHandler) Update(c echo.Context) error {
	var req struct {
		Name          *string        `json:"name"`
		Description   *string        `json:"description"`
		Address       *string        `json:"address"`
		Phone         *string        `json:"phone"`
		BusinessHours map[string]any `json:"business_hours"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	hoursBytes, _ := json.Marshal(req.BusinessHours)
	salon, err := h.svc.UpdateSalon(c.Request().Context(), middleware.GetUserID(c), c.Param("id"), service.UpdateSalonRequest{
		Name:          req.Name,
		Description:   req.Description,
		Address:       req.Address,
		Phone:         req.Phone,
		BusinessHours: hoursBytes,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"salon": salon})
}

// POST /api/v1/salons/:id/portfolio
func (h *SalonHandler) AddPortfolio(c echo.Context) error {
	var req struct {
		ImageURL string `json:"image_url" validate:"required,url"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := h.svc.AddPortfolio(c.Request().Context(), middleware.GetUserID(c), c.Param("id"), req.ImageURL); err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"message": "portfolio image added"})
}
