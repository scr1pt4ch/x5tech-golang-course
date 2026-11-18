package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/lesson9practice/internal/model"
	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

type Orders interface {
	Create(string) (model.Order, error)
	ByID(int) (model.Order, error)
	List(limit, offset int) ([]model.Order, error)
}
type API struct {
	orders Orders
	logger *slog.Logger
}

func New(orders Orders, logger *slog.Logger) *API { return &API{orders: orders, logger: logger} }

// status godoc
// @Summary  Проверка, что сервис работает
// @Tags     status
// @Produce  json
// @Success  200  {object}  StatusResponse
// @Failure  405  {object}  ErrorResponse  "METHOD_NOT_ALLOWED"
// @Router   /status [get]
func (a *API) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, StatusResponse{Status: "ok"})
}

// createOrder godoc
// @Summary      Создать заказ
// @Description  Тело - один JSON-объект с полем customer, не больше 1024 байт.
// @Description  Пробелы по краям имени обрезаются. Сервер сам назначает id и status=new.
// @Description  Повторные заказы одного клиента разрешены.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request  body      CreateOrderRequest  true  "Данные заказа"
// @Success      201      {object}  OrderResponse
// @Header       201      {string}  Location  "Адрес нового заказа, например /api/v1/orders/1"
// @Failure      400      {object}  ErrorResponse  "INVALID_JSON или VALIDATION_ERROR"
// @Failure      405      {object}  ErrorResponse  "METHOD_NOT_ALLOWED"
// @Failure      413      {object}  ErrorResponse  "BODY_TOO_LARGE"
// @Failure      415      {object}  ErrorResponse  "UNSUPPORTED_MEDIA_TYPE"
// @Failure      500      {object}  ErrorResponse  "INTERNAL_ERROR"
// @Router       /orders [post]
func (a *API) createOrder(w http.ResponseWriter, r *http.Request) {
	// Указатель позволяет отличить null от объекта. Теги DTO сами не валидируют запрос.
	var in *CreateOrderRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, err.status, err.message, err.code)
		return
	}
	if in == nil {
		writeError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	order, err := a.orders.Create(in.Customer)
	if errors.Is(err, service.ErrEmptyCustomer) {
		writeError(w, 400, "customer is required", "VALIDATION_ERROR")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/orders/"+strconv.Itoa(order.ID))
	writeJSON(w, 201, orderResponse(order))
}

// getOrder godoc
// @Summary  Получить заказ по номеру
// @Tags     orders
// @Produce  json
// @Param    id   path      int  true  "Номер заказа"  minimum(1)
// @Success  200  {object}  OrderResponse
// @Failure  400  {object}  ErrorResponse  "INVALID_ID"
// @Failure  404  {object}  ErrorResponse  "ORDER_NOT_FOUND"
// @Failure  405  {object}  ErrorResponse  "METHOD_NOT_ALLOWED"
// @Failure  500  {object}  ErrorResponse  "INTERNAL_ERROR"
// @Router   /orders/{id} [get]
func (a *API) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, 400, "id must be a positive integer", "INVALID_ID")
		return
	}
	order, err := a.orders.ByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, 404, "order not found", "ORDER_NOT_FOUND")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	writeJSON(w, 200, orderResponse(order))
}

// listOrders godoc
// @Summary      Список заказов с пагинацией
// @Description  Заказы отсортированы по id по возрастанию. Пустая выборка и страница за концом списка - [].
// @Description  Числа передаются только десятичными цифрами. Неверное, пустое, повторное или неизвестное
// @Description  значение параметра - 400 INVALID_QUERY (значение по умолчанию вместо него не подставляется).
// @Tags         orders
// @Produce      json
// @Param        limit   query     int  false  "Размер страницы"           minimum(1)  maximum(100)  default(20)
// @Param        offset  query     int  false  "Сколько заказов пропустить"  minimum(0)  default(0)
// @Success      200     {array}   OrderResponse
// @Failure      400     {object}  ErrorResponse  "INVALID_QUERY"
// @Failure      405     {object}  ErrorResponse  "METHOD_NOT_ALLOWED"
// @Failure      500     {object}  ErrorResponse  "INTERNAL_ERROR"
// @Router       /orders [get]
func (a *API) listOrders(w http.ResponseWriter, r *http.Request) {
	p, err := parsePage(r.URL.RawQuery)
	if nil != err {
		writeError(w, 400, err.Error(), "INVALID_QUERY")
		return
	}

	orders, err := a.orders.List(p.limit, p.offset)
	if nil != err {
		a.internalError(w, r, err)
		return
	}

	out := make([]OrderResponse, 0, len(orders))
	for _, order := range orders {
		out = append(out, orderResponse(order))
	}
	writeJSON(w, 200, out)
}

func (a *API) internalError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.ErrorContext(r.Context(), "order operation failed", "error", err)
	writeError(w, 500, "internal server error", "INTERNAL_ERROR")
}
