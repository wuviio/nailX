package handler

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type ARHandler struct {
	svc     service.ARService
	cdnBase string
}

func NewARHandler(svc service.ARService, cdnBase string) *ARHandler {
	return &ARHandler{svc: svc, cdnBase: cdnBase}
}

// POST /api/v1/ar/sessions
func (h *ARHandler) CreateSession(c echo.Context) error {
	var req struct {
		DesignIPID           *string  `json:"design_ip_id"`
		DetectedNailLengthMM *float64 `json:"detected_nail_length_mm"`
		HasExistingGel       *bool    `json:"has_existing_gel"`
		DetectedNailShape    *string  `json:"detected_nail_shape"`
		HandSnapshotURL      *string  `json:"hand_snapshot_url"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}

	// api_spec.md Security Note: hand_snapshot_url は CDN 管理下のパスのみ受け付ける
	if req.HandSnapshotURL != nil && *req.HandSnapshotURL != "" {
		if !strings.HasPrefix(*req.HandSnapshotURL, h.cdnBase+"/ar/") {
			return domainErrToHTTP(c, domain.ErrInvalidMediaURL)
		}
	}

	session, err := h.svc.CreateSession(c.Request().Context(), middleware.GetUserID(c), service.CreateARSessionRequest{
		DesignIPID:           req.DesignIPID,
		DetectedNailLengthMM: req.DetectedNailLengthMM,
		HasExistingGel:       req.HasExistingGel,
		DetectedNailShape:    req.DetectedNailShape,
		HandSnapshotURL:      req.HandSnapshotURL,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"ar_session": session})
}

// GET /api/v1/ar/sessions/:id
func (h *ARHandler) GetSession(c echo.Context) error {
	session, err := h.svc.GetSession(c.Request().Context(), middleware.GetUserID(c), c.Param("id"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"ar_session": session})
}
