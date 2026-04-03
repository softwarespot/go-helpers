package middleware

import (
	"net/http"
	"time"

	"github.com/softwarespot/go-helpers/logging"
	"github.com/softwarespot/go-helpers/service"
)

func NewMetrics(logger logging.Logger) service.MiddlewareFunc {
	logger.Log("loaded metrics middleware", logging.LevelNotice)

	return func(next service.Handler) service.Handler {
		return service.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			t0 := time.Now()
			err := next.ServeHTTP(w, r)
			logging.Memory(logger, "handled request",
				service.Args(r,
					"took", time.Since(t0).String(),
					"error", err,
				),
			)
			return err
		})
	}
}
