package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// newRouter пишет логи в консоль
func newRouter(store *Store) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return newRouterWithLogger(store, logger)
}

// newRouterWithLogger нужен для тестов
func newRouterWithLogger(store *Store, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(requestIDMiddleware)
	r.Use(loggingMiddleware(logger))
	r.Use(recoveryMiddleware(logger))

	r.Get("/health", health)
	r.Post("/orders", createOrder(store))
	r.Get("/orders/{id}", getOrder(store))

	return r
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createOrderRequest struct {
	Customer string `json:"customer"`
}

func createOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var req createOrderRequest
		if err := decoder.Decode(&req); nil != err {
			var tooBig *http.MaxBytesError
			if errors.As(err, &tooBig) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if err := decoder.Decode(&struct{}{}); io.EOF != err {
			writeError(w, http.StatusBadRequest, "body must contain only one JSON object")
			return
		}

		customer := strings.TrimSpace(req.Customer)
		if "" == customer {
			writeError(w, http.StatusBadRequest, "customer must not be empty")
			return
		}

		order := store.Create(customer)
		w.Header().Set("Location", "/orders/"+strconv.Itoa(order.ID))
		writeJSON(w, http.StatusCreated, order)
	}
}

func getOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if nil != err || id <= 0 {
			writeError(w, http.StatusBadRequest, "id must be a positive integer")
			return
		}

		order, ok := store.ByID(id)
		if !ok {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeJSON(w, http.StatusOK, order)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if nil != err {
		log.Printf("encode response: %v", err)
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); nil != err {
		log.Printf("write response: %v", err)
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	server := &http.Server{
		Addr: "127.0.0.1:8080", Handler: newRouterWithLogger(NewStore(), logger),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
