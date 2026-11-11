package handler

import (
	"log/slog"
	"net/http"
)

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: capture status code and duration, generate request_id,
			// and write method, path, status, duration and request_id to slog.
			logger.Debug("request received", slog.String("method", r.Method), slog.String("path", r.URL.Path))
			next.ServeHTTP(w, r)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: recover from panic, log the incident and return a safe JSON 500 response.
			next.ServeHTTP(w, r)
		})
	}
}
