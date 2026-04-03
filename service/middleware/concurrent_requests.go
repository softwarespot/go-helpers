package middleware

import (
	"errors"
	"net/http"
	"sync/atomic"

	"github.com/softwarespot/go-helpers/logging"
	"github.com/softwarespot/go-helpers/service"
)

func NewConcurrentRequests(maxConcurrency int32, logger logging.Logger) service.MiddlewareFunc {
	logger.Log("loaded concurrent requests middleware", logging.LevelNotice,
		"max-concurrency", maxConcurrency,
	)

	var concurrency atomic.Int32
	return func(next service.Handler) service.Handler {
		return service.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			defer concurrency.Add(-1)
			if concurrency.Add(1) > maxConcurrency {
				return service.NewError(errors.New(http.StatusText(http.StatusServiceUnavailable)), http.StatusServiceUnavailable)
			}
			return next.ServeHTTP(w, r)
		})
	}
}
