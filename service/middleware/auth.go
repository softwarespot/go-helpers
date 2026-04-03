package middleware

import (
	"context"
	"net/http"

	"github.com/softwarespot/go-helpers/logging"
	"github.com/softwarespot/go-helpers/service"
)

// IMPORTANT: This is an example only
func NewAuthentication(logger logging.Logger) service.MiddlewareFunc {
	type userID string
	logger.Log("loaded authentication middleware", logging.LevelNotice)

	return func(next service.Handler) service.Handler {
		return service.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			// IMPORTANT: Update this part
			ctx := context.WithValue(r.Context(), userID("user"), map[string]any{
				"userId": r.URL.Query().Get("userId"),
			})
			r = r.WithContext(ctx)
			return next.ServeHTTP(w, r)
		})
	}
}
