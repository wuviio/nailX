package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type ReviewHandler struct {
	svc     service.ReviewService
	cdnBase string
}

func NewReviewHandler(svc service.ReviewService, cdnBase string) *ReviewHandler {
	return &ReviewHandler{svc: svc, cdnBase: cdnBase}
}

// POST /api/v1/reviews
func (h *ReviewHandler) Post(c echo.Context) error {
	var req struct {
		BookingID         string  `json:"booking_id" validate:"required"`
		ReproductionScore int     `json:"reproduction_score" validate:"required,min=1,max=5"`
		OverallScore      int     `json:"overall_score" validate:"required,min=1,max=5"`
		Comment           *string `json:"comment"`
		BeforePhotoURL    *string `json:"before_photo_url"`
		AfterPhotoURL     *string `json:"after_photo_url"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}

	// api_spec.md Security Note: 写真 URL は CDN 管理下のパスのみ受け付ける
	reviewPrefix := h.cdnBase + "/reviews/"
	if req.BeforePhotoURL != nil && *req.BeforePhotoURL != "" {
		if !strings.HasPrefix(*req.BeforePhotoURL, reviewPrefix) {
			return domainErrToHTTP(c, domain.ErrInvalidMediaURL)
		}
	}
	if req.AfterPhotoURL != nil && *req.AfterPhotoURL != "" {
		if !strings.HasPrefix(*req.AfterPhotoURL, reviewPrefix) {
			return domainErrToHTTP(c, domain.ErrInvalidMediaURL)
		}
	}

	review, err := h.svc.PostReview(c.Request().Context(), middleware.GetUserID(c), service.PostReviewInput{
		BookingID:         req.BookingID,
		ReproductionScore: req.ReproductionScore,
		OverallScore:      req.OverallScore,
		Comment:           req.Comment,
		BeforePhotoURL:    req.BeforePhotoURL,
		AfterPhotoURL:     req.AfterPhotoURL,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"review": review})
}

// GET /api/v1/salons/:id/reviews
func (h *ReviewHandler) ListBySalon(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 20
	}
	reviews, nextCur, err := h.svc.ListSalonReviews(c.Request().Context(), c.Param("id"), c.QueryParam("cursor"), limit)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"reviews": reviews, "next_cursor": nextCur})
}
