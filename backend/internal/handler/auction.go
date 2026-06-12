package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type AuctionHandler struct {
	svc    service.AuctionService
	sseHub *SSEHub
}

func NewAuctionHandler(svc service.AuctionService, hub *SSEHub) *AuctionHandler {
	return &AuctionHandler{svc: svc, sseHub: hub}
}

// POST /api/v1/auctions/requests
func (h *AuctionHandler) CreateRequest(c echo.Context) error {
	var req struct {
		DesignIPID       string         `json:"design_ip_id" validate:"required"`
		ARSessionID      *string        `json:"ar_session_id"`
		NailDataSnapshot map[string]any `json:"nail_data_snapshot" validate:"required"`
		BudgetMaxYen     int            `json:"budget_max_yen" validate:"required,min=1"`
		DesiredDateFrom  string         `json:"desired_date_from" validate:"required"`
		DesiredDateTo    string         `json:"desired_date_to" validate:"required"`
		AreaPrefecture   string         `json:"area_prefecture" validate:"required"`
		AreaCity         *string        `json:"area_city"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	snapshotBytes, _ := json.Marshal(req.NailDataSnapshot)
	bookingReq, err := h.svc.CreateBookingRequest(c.Request().Context(), middleware.GetUserID(c), service.CreateBookingRequestInput{
		DesignIPID:       req.DesignIPID,
		ARSessionID:      req.ARSessionID,
		NailDataSnapshot: snapshotBytes,
		BudgetMaxYen:     req.BudgetMaxYen,
		DesiredDateFrom:  req.DesiredDateFrom,
		DesiredDateTo:    req.DesiredDateTo,
		AreaPrefecture:   req.AreaPrefecture,
		AreaCity:         req.AreaCity,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"booking_request": bookingReq})
}

// GET /api/v1/auctions/requests/:id
func (h *AuctionHandler) GetRequest(c echo.Context) error {
	req, bids, err := h.svc.GetBookingRequest(c.Request().Context(), middleware.GetUserID(c), c.Param("id"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"booking_request": req, "bids": bids})
}

// DELETE /api/v1/auctions/requests/:id
func (h *AuctionHandler) CancelRequest(c echo.Context) error {
	if err := h.svc.CancelBookingRequest(c.Request().Context(), middleware.GetUserID(c), c.Param("id")); err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// GET /api/v1/auctions/requests (Salon)
func (h *AuctionHandler) ListMatchingRequests(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 20
	}
	reqs, nextCur, err := h.svc.ListMatchingRequests(c.Request().Context(), middleware.GetUserID(c), c.QueryParam("cursor"), limit)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"booking_requests": reqs, "next_cursor": nextCur})
}

// POST /api/v1/auctions/requests/:request_id/bids
func (h *AuctionHandler) PlaceBid(c echo.Context) error {
	var req struct {
		PriceYen             int     `json:"price_yen" validate:"required,min=1"`
		IncludesRemoval      bool    `json:"includes_removal"`
		RemovalFeeYen        int     `json:"removal_fee_yen"`
		AvailableSlotAt      string  `json:"available_slot_at" validate:"required"`
		DynamicDiscountReason *string `json:"dynamic_discount_reason"`
		Message              *string `json:"message"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	bid, err := h.svc.PlaceBid(c.Request().Context(), middleware.GetUserID(c), c.Param("request_id"), service.PlaceBidInput{
		PriceYen:             req.PriceYen,
		IncludesRemoval:      req.IncludesRemoval,
		RemovalFeeYen:        req.RemovalFeeYen,
		AvailableSlotAt:      req.AvailableSlotAt,
		DynamicDiscountReason: req.DynamicDiscountReason,
		Message:              req.Message,
	})
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	// 入札成功時に SSE ブロードキャスト
	_ = h.sseHub.Publish(c.Request().Context(), c.Param("request_id"), &domain.BidEvent{
		Type: "new_bid",
		Bid:  bid,
	})
	return c.JSON(http.StatusCreated, map[string]any{"bid": bid})
}

// PATCH /api/v1/auctions/requests/:request_id/bids/:bid_id
func (h *AuctionHandler) UpdateBid(c echo.Context) error {
	var req struct {
		PriceYen int `json:"price_yen" validate:"required,min=1"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	bid, err := h.svc.UpdateBid(c.Request().Context(), middleware.GetUserID(c), c.Param("request_id"), c.Param("bid_id"), req.PriceYen)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"bid": bid})
}

// GET /api/v1/auctions/requests/:request_id/bids
func (h *AuctionHandler) ListBids(c echo.Context) error {
	bids, err := h.svc.ListBids(c.Request().Context(), middleware.GetUserID(c), c.Param("request_id"), c.QueryParam("sort"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"bids": bids})
}
