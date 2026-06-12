package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/domain"
)

// errResp は api_spec.md の共通エラー形式でレスポンスを返す
func errResp(c echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// domainErrToHTTP はドメインエラーを HTTP ステータスに変換する
func domainErrToHTTP(c echo.Context, err error) error {
	switch err {
	case domain.ErrNotFound:
		return errResp(c, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case domain.ErrUnauthorized:
		return errResp(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
	case domain.ErrForbidden:
		return errResp(c, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case domain.ErrAlreadyBid:
		return errResp(c, http.StatusConflict, "ALREADY_BID", "already bid on this request")
	case domain.ErrIPRejected:
		return errResp(c, http.StatusConflict, "IP_REJECTED", "design rejected as plagiarism")
	case domain.ErrRequestExpired:
		return errResp(c, http.StatusConflict, "REQUEST_EXPIRED", "booking request has expired")
	case domain.ErrRequestLimitReached:
		return errResp(c, http.StatusConflict, "REQUEST_LIMIT_EXCEEDED", "concurrent open request limit of 3 reached")
	case domain.ErrInvalidMediaURL:
		return errResp(c, http.StatusBadRequest, "INVALID_MEDIA_URL", "url must point to the authorized CDN")
	case domain.ErrBudgetExceeded:
		return errResp(c, http.StatusUnprocessableEntity, "BUDGET_EXCEEDED", "bid price exceeds budget")
	case domain.ErrRebidLimitReached:
		return errResp(c, http.StatusConflict, "REBID_LIMIT_REACHED", "rebid limit of 1 reached")
	case domain.ErrInvalidRequest:
		return errResp(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid request parameters")
	default:
		return errResp(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

// cursor はカーソルベースページネーションの next_cursor を返す（レコードなしなら空文字）
func nextCursor(items any, cursor string) string {
	return cursor
}
