package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(recoveryMiddleware(logger))
	router.Use(requestLoggingMiddleware(logger))

	router.Get("/health", h.health)
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.notImplemented)
		r.Get("/", h.notImplemented)
		r.Get("/{id}", h.notImplemented)
		r.Patch("/{id}/status", h.notImplemented)
		r.Get("/{id}/history", h.notImplemented)
	})

	return router
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) notImplemented(w http.ResponseWriter, _ *http.Request) {
	// TODO: replace this placeholder with separate handlers for every endpoint.
	writeJSON(w, http.StatusNotImplemented, errorResponse{
		Error: "endpoint is not implemented",
		Code:  "NOT_IMPLEMENTED",
	})
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
