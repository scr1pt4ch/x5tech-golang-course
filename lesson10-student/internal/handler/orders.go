package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

// createOrder godoc
// @Summary      Создать заказ
// @Description  Сумма заказа считается на сервере, статус нового заказа — new.
// @Description  Названия позиций сохраняются без пробелов по краям.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        input  body      order.CreateOrderInput  true  "Позиции заказа"
// @Success      201    {object}  order.Order
// @Header       201    {string}  Location  "Адрес созданного заказа, например /api/v1/orders/1"
// @Failure      400    {object}  errorResponse  "INVALID_JSON или INVALID_ORDER"
// @Failure      500    {object}  errorResponse  "INTERNAL_ERROR"
// @Router       /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput

	if !readBody(w, r, &input) {
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

// listOrders godoc
// @Summary  Список всех заказов
// @Tags     orders
// @Produce  json
// @Success  200  {array}   order.Order
// @Failure  500  {object}  errorResponse  "INTERNAL_ERROR"
// @Router   /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.List(r.Context())

	if nil != err {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, orders)
}

// getOrder godoc
// @Summary  Получить заказ по id
// @Tags     orders
// @Produce  json
// @Param    id   path      int  true  "Номер заказа"  minimum(1)
// @Success  200  {object}  order.Order
// @Failure  400  {object}  errorResponse  "INVALID_ID"
// @Failure  404  {object}  errorResponse  "ORDER_NOT_FOUND"
// @Failure  500  {object}  errorResponse  "INTERNAL_ERROR"
// @Router   /api/v1/orders/{id} [get]
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromURL(w, r)

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

// updateStatus godoc
// @Summary      Изменить статус заказа
// @Description  Разрешены переходы: new → processing, new → cancelled,
// @Description  processing → completed, processing → cancelled.
// @Description  Из completed и cancelled перейти никуда нельзя.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id     path      int                      true  "Номер заказа"  minimum(1)
// @Param        input  body      order.UpdateStatusInput  true  "Новый статус"
// @Success      200    {object}  order.Order
// @Failure      400    {object}  errorResponse  "INVALID_ID, INVALID_JSON или INVALID_STATUS"
// @Failure      404    {object}  errorResponse  "ORDER_NOT_FOUND"
// @Failure      409    {object}  errorResponse  "INVALID_STATUS_TRANSITION"
// @Failure      500    {object}  errorResponse  "INTERNAL_ERROR"
// @Router       /api/v1/orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromURL(w, r)

	if !ok {
		return
	}

	var input order.UpdateStatusInput

	if !readBody(w, r, &input) {
		return
	}

	updated, err := h.service.UpdateStatus(r.Context(), id, input.Status)

	if nil != err {
		h.writeServiceError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// getHistory godoc
// @Summary      История статусов заказа
// @Description  Записи идут в порядке изменения, первая — new.
// @Tags         orders
// @Produce      json
// @Param        id   path      int  true  "Номер заказа"  minimum(1)
// @Success      200  {array}   order.StatusChange
// @Failure      400  {object}  errorResponse  "INVALID_ID"
// @Failure      404  {object}  errorResponse  "ORDER_NOT_FOUND"
// @Failure      500  {object}  errorResponse  "INTERNAL_ERROR"
// @Router       /api/v1/orders/{id}/history [get]
func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := idFromURL(w, r)

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

func idFromURL(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if nil != err || id <= 0 {
		writeError(w, http.StatusBadRequest, "id must be a positive integer", "INVALID_ID")
		return 0, false
	}

	return id, true
}

func readBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); nil != err {
		writeError(w, http.StatusBadRequest, "request body must be valid JSON", "INVALID_JSON")
		return false
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain one JSON object", "INVALID_JSON")
		return false
	}

	return true
}

var errorCodes = []struct {
	target  error
	status  int
	code    string
	message string
}{
	{order.ErrInvalidOrder, http.StatusBadRequest, "INVALID_ORDER", ""},
	{order.ErrInvalidStatus, http.StatusBadRequest, "INVALID_STATUS", ""},
	{order.ErrOrderNotFound, http.StatusNotFound, "ORDER_NOT_FOUND", "order not found"},
	{order.ErrInvalidStatusTransition, http.StatusConflict, "INVALID_STATUS_TRANSITION", ""},
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	for _, e := range errorCodes {
		if errors.Is(err, e.target) {
			message := e.message
			if "" == message {
				message = err.Error()
			}

			writeError(w, e.status, message, e.code)
			return
		}
	}

	h.logger.Error("unexpected error", "method", r.Method, "path", r.URL.Path, "error", err)
	writeError(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
}
