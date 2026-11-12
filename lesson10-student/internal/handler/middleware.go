package handler

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"
)

const requestIDHeader = "X-Request-ID"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := r.Header.Get(requestIDHeader)
			if "" == requestID {
				requestID = newRequestID()
			}
			w.Header().Set(requestIDHeader, requestID)

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			durationMs := float64(time.Since(start).Microseconds()) / 1000

			level := slog.LevelInfo
			if http.StatusInternalServerError <= rec.status {
				level = slog.LevelError
			}

			logger.Log(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Float64("duration_ms", durationMs),
				slog.String("request_id", requestID),
			)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				panicValue := recover()
				if nil == panicValue {
					return
				}

				logger.Error("panic recovered",
					slog.Any("panic", panicValue),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("request_id", w.Header().Get(requestIDHeader)),
				)

				writeError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func newRequestID() string {
	return fmt.Sprintf("%016x", rand.Uint64())
}
