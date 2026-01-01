package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/softwarespot/go-helpers/logging"
)

type ResponseWriter struct {
	logger logging.Logger
}

func NewResponseWriter(logger logging.Logger) *ResponseWriter {
	return &ResponseWriter{
		logger: logger,
	}
}

func (rw *ResponseWriter) WriteAsJSON(w http.ResponseWriter, _ *http.Request, res any) error {
	// See URL: https://journal.petrausch.info/post/2020/06/golang-json-encoder-http-response-writer/,
	// which explains why an error can be returned
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		return fmt.Errorf("unable to encode the HTTP application/json response: %w", err)
	}
	return nil
}

func (rw *ResponseWriter) ErrorAsJSON(w http.ResponseWriter, r *http.Request, err error, args ...any) {
	errMsg, statusCode := getErrorStatus(err)
	args = append([]any{"status-code", statusCode}, args...)
	rw.logger.LogError(err, logging.LevelError, Args(r, args)...)

	// Similar to "http.Error()"
	h := w.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", "application/json")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)

	res := map[string]any{
		"msg":  errMsg,
		"code": statusCode,
		"url":  r.URL.RequestURI(),
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		err = fmt.Errorf("unable to encode the HTTP error response: %w", err)
		rw.logger.LogError(err, logging.LevelError, Args(r, args)...)
	}
}

func (rw *ResponseWriter) Error(w http.ResponseWriter, r *http.Request, err error, args ...any) {
	errMsg, statusCode := getErrorStatus(err)
	args = append(
		[]any{
			"status-code", statusCode,
		},
		args...,
	)
	rw.logger.LogError(err, logging.LevelError, Args(r, args)...)
	w.WriteHeader(statusCode)
	http.Error(w, errMsg, statusCode)
}

func getErrorStatus(err error) (string, int) {
	var e Error
	if errors.As(err, &e) {
		return e.Unwrap().Error(), e.Status()
	}
	return http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError
}
