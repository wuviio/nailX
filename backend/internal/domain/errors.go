package domain

import "errors"

var (
	ErrNotFound              = errors.New("not_found")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrForbidden             = errors.New("forbidden")
	ErrAlreadyExists         = errors.New("already_exists")
	ErrAlreadyBid            = errors.New("already_bid")
	ErrIPRejected            = errors.New("ip_rejected")
	ErrRequestExpired        = errors.New("request_expired")
	ErrRequestLimitReached   = errors.New("request_limit_reached")
	ErrBudgetExceeded        = errors.New("budget_exceeded")
	ErrRebidLimitReached     = errors.New("rebid_limit_reached")
	ErrInvalidRequest        = errors.New("invalid_request")
	ErrInvalidMediaURL       = errors.New("invalid_media_url")
)
