package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRouter создаёт роутер с пустым хранилищем
// Логи пишутся в буфер, чтобы тест мог их прочитать
func newTestRouter() (*chi.Mux, *bytes.Buffer) {
	logs := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(logs, nil))
	return newRouterWithLogger(NewStore(), logger), logs
}

// sendRequest отправляет запрос в роутер и возвращает ответ
func sendRequest(h http.Handler, method, path, body, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if "" != contentType {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// checkError проверяет статус, Content-Type и поле error в ответе
func checkError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	require.Equal(t, wantStatus, rec.Code, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body.Error)
}

// checkOrder проверяет статус, Content-Type и поля заказа в ответе
func checkOrder(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, want Order) {
	t.Helper()
	require.Equal(t, wantStatus, rec.Code, rec.Body.String())
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got Order
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, want, got)
}

func TestCreateOrder(t *testing.T) {
	router, _ := newTestRouter()

	rec := sendRequest(router, "POST", "/orders", `{"customer":"Анна"}`, "application/json")
	checkOrder(t, rec, 201, Order{ID: 1, Customer: "Анна", Status: "new"})
	assert.Equal(t, "/orders/1", rec.Header().Get("Location"))

	// Второй заказ получает следующий ID.
	rec = sendRequest(router, "POST", "/orders", `{"customer":"Борис"}`, "application/json")
	checkOrder(t, rec, 201, Order{ID: 2, Customer: "Борис", Status: "new"})
	assert.Equal(t, "/orders/2", rec.Header().Get("Location"))
}

func TestGetCreatedOrder(t *testing.T) {
	router, _ := newTestRouter()
	sendRequest(router, "POST", "/orders", `{"customer":"Анна"}`, "application/json")

	rec := sendRequest(router, "GET", "/orders/1", "", "")
	checkOrder(t, rec, 200, Order{ID: 1, Customer: "Анна", Status: "new"})
}

func TestGetOrderErrors(t *testing.T) {
	router, _ := newTestRouter()

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"неизвестный заказ", "/orders/999", 404},
		{"не число", "/orders/abc", 400},
		{"ноль", "/orders/0", 400},
		{"отрицательный", "/orders/-5", 400},
		{"слишком большое число", "/orders/99999999999999999999", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := sendRequest(router, "GET", tt.path, "", "")
			checkError(t, rec, tt.wantStatus)
		})
	}
}

func TestCreateOrderErrors(t *testing.T) {
	router, _ := newTestRouter()

	tests := []struct {
		name        string
		body        string
		contentType string
		wantStatus  int
	}{
		{"пустое тело", "", "application/json", 400},
		{"сломанный JSON", `{"customer":`, "application/json", 400},
		{"customer число", `{"customer":123}`, "application/json", 400},
		{"пустое имя", `{"customer":""}`, "application/json", 400},
		{"имя из пробелов", `{"customer":"   "}`, "application/json", 400},
		{"клиент прислал id", `{"customer":"Анна","id":10}`, "application/json", 400},
		{"клиент прислал status", `{"customer":"Анна","status":"done"}`, "application/json", 400},
		{"два JSON-объекта", `{"customer":"Анна"}{"customer":"Борис"}`, "application/json", 400},
		{"текстовый Content-Type", `{"customer":"Анна"}`, "text/plain", 415},
		{"без Content-Type", `{"customer":"Анна"}`, "", 415},
		{"слишком большое тело", `{"customer":"` + strings.Repeat("a", 2000) + `"}`, "application/json", 413},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := sendRequest(router, "POST", "/orders", tt.body, tt.contentType)
			checkError(t, rec, tt.wantStatus)
		})
	}

	rec := sendRequest(router, "POST", "/orders", `{"customer":"Анна"}`, "application/json")
	checkOrder(t, rec, 201, Order{ID: 1, Customer: "Анна", Status: "new"})
}

func TestRequestIDInLog(t *testing.T) {
	router, logs := newTestRouter()

	rec := sendRequest(router, "POST", "/orders", `{"customer":"Анна"}`, "application/json")

	id := rec.Header().Get("X-Request-ID")
	require.NotEmpty(t, id)

	log := logs.String()
	assert.Contains(t, log, `"request_id":"`+id+`"`)
	assert.Contains(t, log, `"method":"POST"`)
	assert.Contains(t, log, `"path":"/orders"`)
	assert.Contains(t, log, `"status":201`)
	assert.Contains(t, log, `"duration_ms":`)

	assert.NotContains(t, log, "Анна")
}

func TestPanicReturns500(t *testing.T) {
	router, logs := newTestRouter()

	router.Get("/test-panic", func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	rec := sendRequest(router, "GET", "/test-panic", "", "")
	checkError(t, rec, 500)

	log := logs.String()
	id := rec.Header().Get("X-Request-ID")

	assert.Equal(t, 2, strings.Count(log, `"request_id":"`+id+`"`), log)
	assert.Contains(t, log, `"status":500`)

	assert.Equal(t, 200, sendRequest(router, "GET", "/health", "", "").Code)
}
