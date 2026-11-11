package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput

	if !decodeJSON(w, r, &input) {
		return
	}

	created, err := h.service.Create(r.Context(), input)

	if nil != err {
		h.writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Location", "/api/v1/orders/"+strconv.FormatInt(created.ID, 10))
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())

	if nil != err {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)

	if !ok {
		return
	}

	found, err := h.service.GetByID(r.Context(), id)

	if nil != err {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, found)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)

	if !ok {
		return
	}

	var input order.UpdateStatusInput

	if !decodeJSON(w, r, &input) {
		return
	}

	updated, err := h.service.UpdateStatus(r.Context(), id, input.Status)

	if nil != err {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)

	if !ok {
		return
	}

	history, err := h.service.History(r.Context(), id)

	if nil != err {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, history)
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if nil != err || id <= 0 {
		writeError(w, http.StatusBadRequest, "id <= 0", "INVALID_ID")

		return 0, false
	}

	return id, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); nil != err {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON", "INVALID_JSON")
		return false
	}

	return true
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, order.ErrInvalidOrder):
		writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
	case errors.Is(err, order.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, err.Error(), "INVALID_STATUS")
	case errors.Is(err, order.ErrOrderNotFound):
		writeError(w, http.StatusNotFound, "order not found", "ORDER_NOT_FOUND")
	case errors.Is(err, order.ErrInvalidStatusTransition):
		writeError(w, http.StatusConflict, err.Error(), "INVALID_STATUS_TRANSITION")
	default:
		h.logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
	}
}
