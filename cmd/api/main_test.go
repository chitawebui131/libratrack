package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func do(t *testing.T, r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/health", "")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestUnknownRouteReturns404(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/no-such-route", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestBookRoutesAreRegistered(t *testing.T) {
	r := setupRouter()
	want := map[string]bool{
		"POST /api/v1/books":       false,
		"GET /api/v1/books":        false,
		"GET /api/v1/books/:id":    false,
		"PUT /api/v1/books/:id":    false,
		"DELETE /api/v1/books/:id": false,
		"GET /health":              false,
	}
	for _, ri := range r.Routes() {
		key := ri.Method + " " + ri.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for route, found := range want {
		if !found {
			t.Errorf("route not registered: %s", route)
		}
	}
}

func TestSwaggerIsServed(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/swagger/index.html", "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestListBooksEmpty(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/v1/books", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMethodNotAllowedOnHealth(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/health", "")
	if w.Code == http.StatusOK {
		t.Errorf("POST /health must not return 200")
	}
}

func TestCreateBookInvalidJSON(t *testing.T) {
	w := do(t, setupRouter(), http.MethodPost, "/api/v1/books", `{not json`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestGetBookInvalidID(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/v1/books/abc", "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetBookNotFound(t *testing.T) {
	w := do(t, setupRouter(), http.MethodGet, "/api/v1/books/999", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteBookNotFound(t *testing.T) {
	w := do(t, setupRouter(), http.MethodDelete, "/api/v1/books/999", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/books", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Errorf("expected Access-Control-Allow-Origin header on preflight response")
	}
}