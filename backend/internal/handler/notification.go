package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/middleware"
	"github.com/nailx/backend/internal/service"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// GET /api/v1/notifications
func (h *NotificationHandler) List(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit == 0 {
		limit = 20
	}
	var isRead *bool
	if v := c.QueryParam("is_read"); v != "" {
		b := v == "true"
		isRead = &b
	}
	notifications, nextCur, err := h.svc.ListNotifications(c.Request().Context(), middleware.GetUserID(c), isRead, c.QueryParam("cursor"), limit)
	if err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"notifications": notifications, "next_cursor": nextCur})
}

// PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(c echo.Context) error {
	if err := h.svc.MarkRead(c.Request().Context(), middleware.GetUserID(c), c.Param("id")); err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// POST /api/v1/notifications/fcm-token
func (h *NotificationHandler) RegisterFCMToken(c echo.Context) error {
	var req struct {
		FCMToken       string `json:"fcm_token" validate:"required"`
		DevicePlatform string `json:"device_platform" validate:"required,oneof=ios android"`
	}
	if err := c.Bind(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}
	if err := h.svc.RegisterFCMToken(c.Request().Context(), middleware.GetUserID(c), req.FCMToken, req.DevicePlatform); err != nil {
		return domainErrToHTTP(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"message": "fcm token registered"})
}
