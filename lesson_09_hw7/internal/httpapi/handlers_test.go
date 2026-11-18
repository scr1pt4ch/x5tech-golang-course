package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

func newHandler() http.Handler {
	return New(service.New(repository.New()), slog.New(slog.NewTextHandler(io.Discard, nil))).Handler()
}

func request(handler http.Handler, method, path, body, contentType string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestCreateAndReadOrder(t *testing.T) {
	handler := newHandler()
	created := request(handler, "POST", "/api/v1/orders", `{"customer":"  Алиса  "}`, "application/json")
	if created.Code != 201 {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body)
	}
	if got := created.Header().Get("Location"); got != "/api/v1/orders/1" {
		t.Fatalf("Location=%q", got)
	}
	var order OrderResponse
	if err := json.Unmarshal(created.Body.Bytes(), &order); err != nil {
		t.Fatal(err)
	}
	if order.ID != 1 || order.Customer != "Алиса" || order.Status != "new" {
		t.Fatalf("unexpected order: %+v", order)
	}
	read := request(handler, "GET", created.Header().Get("Location"), "", "")
	if read.Code != 200 || read.Body.String() != created.Body.String() {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body)
	}
	// Повторный заказ одного клиента разрешён: имя не является уникальным ключом.
	again := request(handler, "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
	if again.Code != 201 || again.Header().Get("Location") != "/api/v1/orders/2" {
		t.Fatalf("second create status=%d Location=%q", again.Code, again.Header().Get("Location"))
	}
}

func TestErrorsKeepJSONContract(t *testing.T) {
	tests := []struct {
		name, method, path, body, contentType string
		status                                int
		code                                  string
	}{
		{"invalid id", "GET", "/api/v1/orders/abc", "", "", 400, "INVALID_ID"},
		{"zero id", "GET", "/api/v1/orders/0", "", "", 400, "INVALID_ID"},
		{"missing order", "GET", "/api/v1/orders/123", "", "", 404, "ORDER_NOT_FOUND"},
		{"missing route", "GET", "/api/v1/missing", "", "", 404, "ROUTE_NOT_FOUND"},
		{"wrong method", "PUT", "/api/v1/orders/1", "", "", 405, "METHOD_NOT_ALLOWED"},
		{"empty customer", "POST", "/api/v1/orders", `{"customer":" "}`, "application/json", 400, "VALIDATION_ERROR"},
		{"null body", "POST", "/api/v1/orders", `null`, "application/json", 400, "INVALID_JSON"},
		{"wrong type", "POST", "/api/v1/orders", `{"customer":42}`, "application/json", 400, "INVALID_JSON"},
		{"unknown field", "POST", "/api/v1/orders", `{"customer":"Алиса","status":"paid"}`, "application/json", 400, "INVALID_JSON"},
		{"second JSON", "POST", "/api/v1/orders", `{"customer":"Алиса"} {}`, "application/json", 400, "INVALID_JSON"},
		{"wrong media type", "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "text/plain", 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"large body", "POST", "/api/v1/orders", `{"customer":"` + strings.Repeat("a", 1100) + `"}`, "application/json", 413, "BODY_TOO_LARGE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := request(newHandler(), tt.method, tt.path, tt.body, tt.contentType)
			if w.Code != tt.status {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tt.status, w.Body)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type=%q", got)
			}
			var out ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Code != tt.code || out.Error == "" {
				t.Fatalf("unexpected error: %+v", out)
			}
		})
	}
}

func TestUnsupportedMethodDescribesAllowedMethods(t *testing.T) {
	handler := newHandler()
	for _, tt := range []struct{ method, path, allow string }{
		{"DELETE", "/api/v1/orders/1", "GET, HEAD"},
		{"POST", "/api/v1/status", "GET, HEAD"},
		{"DELETE", "/api/v1/orders", "GET, HEAD, POST"},
	} {
		w := request(handler, tt.method, tt.path, "", "")
		if w.Code != 405 || w.Header().Get("Allow") != tt.allow {
			t.Errorf("%s %s: status=%d Allow=%q", tt.method, tt.path, w.Code, w.Header().Get("Allow"))
		}
	}
}

func TestStatus(t *testing.T) {
	w := request(newHandler(), "GET", "/api/v1/status", "", "")
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("status=%d body=%s", w.Code, w.Body)
	}
}

func createOrders(t *testing.T, handler http.Handler, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		w := request(handler, "POST", "/api/v1/orders", `{"customer":"Клиент"}`, "application/json")
		if 201 != w.Code {
			t.Fatalf("create status=%d body=%s", w.Code, w.Body)
		}
	}
}

func listIDs(t *testing.T, handler http.Handler, query string) []int {
	t.Helper()
	w := request(handler, "GET", "/api/v1/orders"+query, "", "")
	if 200 != w.Code {
		t.Fatalf("list%s status=%d body=%s", query, w.Code, w.Body)
	}
	if got := w.Header().Get("Content-Type"); "application/json" != got {
		t.Fatalf("Content-Type=%q", got)
	}
	var orders []OrderResponse
	if err := json.Unmarshal(w.Body.Bytes(), &orders); nil != err {
		t.Fatal(err)
	}
	ids := make([]int, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.ID)
	}
	return ids
}

func TestListEmptyIsArray(t *testing.T) {
	w := request(newHandler(), "GET", "/api/v1/orders", "", "")
	if 200 != w.Code {
		t.Fatalf("status=%d body=%s", w.Code, w.Body)
	}

	if body := strings.TrimSpace(w.Body.String()); "[]" != body {
		t.Fatalf("body=%s, want []", body)
	}
}

func TestListPages(t *testing.T) {
	handler := newHandler()
	createOrders(t, handler, 3)

	tests := []struct {
		name  string
		query string
		want  []int
	}{
		{"без параметров", "", []int{1, 2, 3}},
		{"первая страница", "?limit=2", []int{1, 2}},
		{"вторая страница", "?limit=2&offset=2", []int{3}},
		{"второй заказ", "?limit=1&offset=1", []int{2}},
		{"limit больше числа заказов", "?limit=100", []int{1, 2, 3}},
		{"offset равен числу заказов", "?offset=3", []int{}},
		{"offset за концом списка", "?offset=50", []int{}},
		{"огромный offset", "?offset=9223372036854775807", []int{}},
		{"огромный offset и limit", "?limit=100&offset=9223372036854775807", []int{}},
		{"ведущие нули", "?limit=001&offset=01", []int{2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listIDs(t, handler, tt.query)
			if len(got) != len(tt.want) {
				t.Fatalf("ids=%v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if tt.want[i] != got[i] {
					t.Fatalf("ids=%v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestListRejectsBadQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{"пустой limit", "?limit="},
		{"пустой offset", "?offset="},
		{"буквы в limit", "?limit=abc"},
		{"буквы в offset", "?offset=1a"},
		{"отрицательный offset", "?offset=-1"},
		{"плюс перед числом", "?limit=+5"},
		{"дробное число", "?limit=1.5"},
		{"пробел в числе", "?limit=%201"},
		{"limit=0", "?limit=0"},
		{"limit=101", "?limit=101"},
		{"limit не помещается в int", "?limit=99999999999999999999"},
		{"offset не помещается в int", "?offset=99999999999999999999"},
		{"повторный limit", "?limit=1&limit=2"},
		{"повторный offset", "?offset=1&offset=1"},
		{"неизвестный параметр", "?page=2"},
		{"параметр без значения", "?limit"},
		{"битая строка запроса", "?limit=%zz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := request(newHandler(), "GET", "/api/v1/orders"+tt.query, "", "")
			if 400 != w.Code {
				t.Fatalf("status=%d body=%s", w.Code, w.Body)
			}
			var out ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &out); nil != err {
				t.Fatal(err)
			}
			if "INVALID_QUERY" != out.Code || "" == out.Error {
				t.Fatalf("unexpected error: %+v", out)
			}
		})
	}
}

func TestSwaggerUI(t *testing.T) {
	handler := newHandler()
	for _, path := range []string{"/swagger/index.html", "/swagger/doc.json"} {
		w := request(handler, "GET", path, "", "")
		if 200 != w.Code {
			t.Errorf("%s: status=%d", path, w.Code)
		}
	}
}
