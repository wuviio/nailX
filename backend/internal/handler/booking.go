package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type BookingHandler struct {
	svc service.BookingService
}

func NewBookingHandler(svc service.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// POST /api/v1/bookings
func (h *BookingHandler) Confirm(c echo.Context) error {
	var req struct {
		BidID           string `json:"bid_id" validate:"required"`
		PaymentMethodID string `json:"payment_method_id" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	booking, payment, clientSecret, err := h.svc.ConfirmBooking(c.Request().Context(), middleware.GetUserID(c), req.BidID, req.PaymentMethodID)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{
		"booking":              booking,
		"payment":              payment,
		"stripe_client_secret": clientSecret,
	})
}

// GET /api/v1/bookings
func (h *BookingHandler) List(c echo.Context) error {
	bookings, err := h.svc.ListBookings(c.Request().Context(), middleware.GetUserID(c), c.QueryParam("status"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"bookings": bookings})
}

// GET /api/v1/bookings/:id
func (h *BookingHandler) Get(c echo.Context) error {
	booking, err := h.svc.GetBooking(c.Request().Context(), middleware.GetUserID(c), c.Param("id"))
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"booking": booking})
}

// POST /api/v1/bookings/:id/complete [Salon]
func (h *BookingHandler) Complete(c echo.Context) error {
	if err := h.svc.CompleteBooking(c.Request().Context(), middleware.GetUserID(c), c.Param("id")); err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "booking completed"})
}

// POST /api/v1/bookings/:id/cancel
func (h *BookingHandler) Cancel(c echo.Context) error {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := h.svc.CancelBooking(c.Request().Context(), middleware.GetUserID(c), c.Param("id"), req.Reason); err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "booking cancelled"})
}
