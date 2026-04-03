package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/softwarespot/go-helpers/logging"
	"github.com/softwarespot/go-helpers/service"
)

type PanicRecoveryFunc func(http.ResponseWriter, *http.Request, error)

func NewPanicRecovery(fn PanicRecoveryFunc, logger logging.Logger) service.MiddlewareFunc {
	logger.Log("loaded panic recovery middleware", logging.LevelNotice)

	return func(next service.Handler) service.Handler {
		return service.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			defer panicRecovery(w, r, fn)
			return next.ServeHTTP(w, r)
		})
	}
}

func panicRecovery(w http.ResponseWriter, r *http.Request, fn PanicRecoveryFunc) {
	if rvr := recover(); rvr != nil {
		var err error
		switch e := rvr.(type) {
		case error:
			err = fmt.Errorf("recovered panic: %w", e)
		default:
			err = fmt.Errorf("%v", e)
		}
		if err == nil {
			err = errors.New("unexpected nil error")
		}
		fn(w, r, err)
	}
}
