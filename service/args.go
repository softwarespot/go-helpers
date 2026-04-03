package service

import "net/http"

func Args(r *http.Request, args ...any) []any {
	return append(
		[]any{
			"url", r.URL.RequestURI(),
			"http-method", r.Method,
			"ip", r.RemoteAddr,
			"real-ip", r.Header.Get("X-Real-IP"),
			"forwarded-ip", r.Header.Get("X-Forwarded-For"),
		},
		args...,
	)
}
