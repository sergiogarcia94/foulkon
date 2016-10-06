package middleware

import "net/http"

const (
	// HTTP Header
	REQUEST_ID_HEADER = "X-Request-Id"
	USER_ID_HEADER    = "X-FOULKON-USER-ID"

	// Middleware names
	AUTHENTICATOR_MIDDLEWARE  = "AUTHENTICATOR"
	XREQUESTID_MIDDLEWARE     = "XREQUESTID"
	REQUEST_LOGGER_MIDDLEWARE = "REQUEST-LOGGER"
)

// MiddlewareHandler is handler with a middleware list to apply
type MiddlewareHandler struct {
	Middlewares map[string]Middleware
}

// MiddlewareContext with all parameters used in the context of middlewares
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

// Handle method execute middlewares in correct order after API handler
func (mwh *MiddlewareHandler) Handle(apiHandler http.Handler) http.Handler {
	var handler http.Handler
	// Wrap target handler to use middleware
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiHandler.ServeHTTP(w, r)
	})
	handler = mwh.Middlewares[REQUEST_LOGGER_MIDDLEWARE].Action(handler)
	handler = mwh.Middlewares[AUTHENTICATOR_MIDDLEWARE].Action(handler)
	handler = mwh.Middlewares[XREQUESTID_MIDDLEWARE].Action(handler)

	return handler
}

// GetMiddlewareContext method retrieves all information about middleware context applied to request
func (mwh *MiddlewareHandler) GetMiddlewareContext(r *http.Request) *MiddlewareContext {
	context := new(MiddlewareContext)
	for _, m := range mwh.Middlewares {
		m.GetInfo(r, context)
	}

	return context
}
