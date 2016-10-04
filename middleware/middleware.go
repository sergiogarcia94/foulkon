package middleware

import (
	"net/http"
)

const (
	// HTTP Header
	REQUEST_ID_HEADER = "X-Request-Id"
	USER_ID_HEADER    = "X-FOULKON-USER-ID"
)

type MiddlewareContext struct {
	// Authenticator middleware
	UserId string
	Admin  bool

	// X-Request-Id middleware
	XRequestId string
}

// Middleware interface with all operations that must be implemented
type Middleware interface {
	// Action to do per each request
	Action(next http.Handler) http.Handler

	// Additional info that middleware use
	GetInfo(r *http.Request, mc *MiddlewareContext)
}
