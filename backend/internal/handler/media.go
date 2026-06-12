package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type MediaHandler struct {
	svc service.MediaService
}

func NewMediaHandler(svc service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

// POST /api/v1/media/presigned-url
// purpose: "ar_snapshot" | "portfolio" | "review_photo" | "design_preview"
func (h *MediaHandler) GeneratePresignedURL(c echo.Context) error {
	var req struct {
		FileType string `json:"file_type" validate:"required"`
		Purpose  string `json:"purpose" validate:"required,oneof=ar_snapshot portfolio review_photo design_preview"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	uploadURL, fileURL, err := h.svc.GeneratePresignedURL(c.Request().Context(), middleware.GetUserID(c), req.FileType, req.Purpose)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"upload_url": uploadURL,
		"file_url":   fileURL,
	})
}
