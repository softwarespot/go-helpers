package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/softwarespot/go-helpers/logging"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return h(w, r)
}

type MiddlewareFunc func(Handler) Handler

type Server struct {
	logger               logging.Logger
	server               *http.Server
	mux                  *http.ServeMux
	middlewares          []MiddlewareFunc
	errHandlersByPattern map[string]func(w http.ResponseWriter, r *http.Request, err error)
}

func NewServer(addr string, logger logging.Logger) *Server {
	mux := http.NewServeMux()
	return &Server{
		logger: logger,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,

			// Fix potential Slowloris Attack because ReadHeaderTimeout is not configured in the http.Server
			ReadHeaderTimeout: 10 * time.Second,
		},
		mux:                  mux,
		middlewares:          nil,
		errHandlersByPattern: make(map[string]func(w http.ResponseWriter, r *http.Request, err error)),
	}
}

// Use adds one or more middleware functions to the server.
// Middleware functions are executed in the order they are added,
// wrapping the handler functions registered with HandleFunc.
// This allows for cross-cutting concerns such as logging, authentication,
// and request timing to be handled consistently across all routes
func (s *Server) Use(middleware ...MiddlewareFunc) {
	s.middlewares = append(s.middlewares, middleware...)
}

// Handle registers the handler for the given pattern.
// If the given pattern conflicts, with one that is already registered, Handle
// panics
func (s *Server) Handle(pattern string, handler Handler) {
	s.HandleFunc(pattern, handler.ServeHTTP)
}

// HandleFunc registers the handler function for the given pattern.
// If the given pattern conflicts, with one that is already registered, HandleFunc
// panics
func (s *Server) HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request) error) {
	h := s.applyMiddlewareHandlers(HandlerFunc(handler))
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if err := h.ServeHTTP(w, r); err != nil {
			if h, ok := s.errHandlersByPattern[pattern]; ok {
				h(w, r, err)
			} else {
				_, statusCode := getErrorStatus(err)
				args := []any{
					"status-code", statusCode,
				}
				s.logger.LogError(err, logging.LevelError, Args(r, args)...)
				w.WriteHeader(statusCode)
			}
		}
	})
}

// HandleErrorFunc registers the error handler function for the given pattern.
// If the given pattern conflicts, with one that is already registered, HandleErrorFunc
// panics
func (s *Server) HandleErrorFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request, err error)) {
	if handler == nil {
		panic("nil error handler")
	}
	if _, ok := s.errHandlersByPattern[pattern]; ok {
		panic("multiple error handler registrations for " + pattern)
	}
	s.errHandlersByPattern[pattern] = handler
}

func (s *Server) applyMiddlewareHandlers(handler Handler) Handler {
	// Ensure the middleware handlers are executed in the same order they were registered i.e. FIFO
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		mwh := s.middlewares[i]
		handler = mwh(handler)
	}
	return handler
}

// ListenAndServe starts the HTTP server (in a separate go routine) and listens for incoming requests.
// If an error occurs while starting the server, it will return that error.
// The method also handles graceful shutdown when the provided context is done
func (s *Server) ListenAndServe(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		// Ignore the error if nil or is a server closed error
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server unexpectedly closed: %w", err)
		}
	}()

	s.logger.Log("started listening", logging.LevelNotice,
		"address", s.server.Addr,
	)

	// Wait for either the context to be done or a non-closing error from the "ListenAndServe()" function
	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	s.logger.Log("start graceful server shutdown", logging.LevelNotice,
		"address", s.server.Addr,
		"context-error", ctx.Err(),
	)
	err := s.server.Shutdown(ctx)
	s.logger.Log("stop graceful server shutdown", logging.LevelNotice,
		"address", s.server.Addr,
		"shutdown-error", err,
	)
	return err
}
