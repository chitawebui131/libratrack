package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack/internal/handler"
	"libratrack/internal/middleware"
	"libratrack/internal/repository"
)

// Request and response keys come from json tags (snake_case): id, title, created_at, etc.
const validBook = `{
	"title": "The Go Programming Language",
	"isbn": "9780134190440",
	"author": "Donovan, Kernighan",
	"category": "programming",
	"published_year": 2015,
	"description": "Classic Go book"
}`

const updatedBook = `{
	"title": "TGPL 2nd printing",
	"isbn": "9780134190440",
	"author": "Donovan, Kernighan",
	"category": "programming",
	"published_year": 2016,
	"description": "Updated"
}`

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware()) // as in main.go: a panic becomes a 500
	h := handler.NewBookHandler(repository.NewInMemoryBookRepository())

	v1 := r.Group("/api/v1")
	v1.POST("/books", h.CreateBook)
	v1.GET("/books", h.ListBooks)
	v1.GET("/books/:id", h.GetBook)
	v1.PUT("/books/:id", h.UpdateBook)
	v1.DELETE("/books/:id", h.DeleteBook)
	return r
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

func errorCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v (body: %s)", err, w.Body.String())
	}
	return resp.Error.Code
}

func TestInvalidID_Returns400(t *testing.T) {
	r := newRouter()
	ids := []string{"abc", "-1", "1.5", "4294967296"} // not a number, negative, fractional, > uint32

	for _, id := range ids {
		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			t.Run(method+"/"+id, func(t *testing.T) {
				w := do(t, r, method, "/api/v1/books/"+id, validBook)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("expected 400, got %d", w.Code)
				}
				if c := errorCode(t, w); c != "invalid_id" {
					t.Errorf("expected error code invalid_id, got %q", c)
				}
			})
		}
	}
}

func TestMalformedJSON_Returns422(t *testing.T) {
	r := newRouter()

	cases := []struct{ name, method, path string }{
		{"create", http.MethodPost, "/api/v1/books"},
		{"update", http.MethodPut, "/api/v1/books/1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := do(t, r, tc.method, tc.path, `{not json`)
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected 422, got %d", w.Code)
			}
			if c := errorCode(t, w); c != "validation_error" {
				t.Errorf("expected validation_error, got %q", c)
			}
		})
	}
}

func TestNotFound_Returns404(t *testing.T) {
	r := newRouter()

	cases := []struct{ name, method, body string }{
		{"get", http.MethodGet, ""},
		{"update", http.MethodPut, validBook},
		{"delete", http.MethodDelete, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := do(t, r, tc.method, "/api/v1/books/999", tc.body)
			if w.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
			}
			if c := errorCode(t, w); c != "not_found" {
				t.Errorf("expected not_found, got %q", c)
			}
		})
	}
}

func TestListBooks_EmptyReturns200(t *testing.T) {
	w := do(t, newRouter(), http.MethodGet, "/api/v1/books", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestListBooks_InvalidPaginationReturns400(t *testing.T) {
	r := newRouter()
	queries := []string{"?page=abc", "?limit=abc", "?page=0", "?limit=0", "?page=-1&limit=-5"}

	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			w := do(t, r, http.MethodGet, "/api/v1/books"+q, "")
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
			if c := errorCode(t, w); c != "invalid_pagination" {
				t.Errorf("expected invalid_pagination, got %q", c)
			}
		})
	}
}

func TestListBooks_EmptyReturnsJSONArray(t *testing.T) {
	w := do(t, newRouter(), http.MethodGet, "/api/v1/books", "")
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Errorf("expected [], got %s", w.Body.String())
	}
}

func createBook(t *testing.T, r http.Handler, title, category string) {
	t.Helper()
	body := mustJSON(t, map[string]any{"title": title, "isbn": "I", "author": "A", "category": category})
	if w := do(t, r, http.MethodPost, "/api/v1/books", body); w.Code != http.StatusCreated {
		t.Fatalf("create %q: expected 201, got %d", title, w.Code)
	}
}

func listTitles(t *testing.T, r http.Handler, query string) []string {
	t.Helper()
	w := do(t, r, http.MethodGet, "/api/v1/books"+query, "")
	if w.Code != http.StatusOK {
		t.Fatalf("list %s: expected 200, got %d", query, w.Code)
	}
	var books []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &books); err != nil {
		t.Fatal(err)
	}
	titles := make([]string, 0, len(books))
	for _, b := range books {
		titles = append(titles, b["title"].(string))
	}
	return titles
}

func TestListBooks_FilterAndPagination(t *testing.T) {
	r := newRouter()
	createBook(t, r, "first", "a")
	createBook(t, r, "second", "b")
	createBook(t, r, "third", "a")

	if got := listTitles(t, r, ""); len(got) != 3 {
		t.Errorf("no filter: expected 3 books, got %v", got)
	}
	if got := listTitles(t, r, "?category=a"); len(got) != 2 {
		t.Errorf("category=a: expected 2 books, got %v", got)
	}
	if got := listTitles(t, r, "?page=2&limit=1"); len(got) != 1 || got[0] != "second" {
		t.Errorf("page=2&limit=1: expected [second], got %v", got)
	}
	if got := listTitles(t, r, "?page=9&limit=1"); len(got) != 0 {
		t.Errorf("page beyond the end: expected no books, got %v", got)
	}
}

func TestListBooks_LimitIsCappedAt100(t *testing.T) {
	r := newRouter()
	for i := 0; i < 105; i++ {
		createBook(t, r, fmt.Sprintf("book-%d", i), "a")
	}

	if got := listTitles(t, r, "?limit=1000"); len(got) != 100 {
		t.Errorf("expected 100 books, got %d", len(got))
	}
}

func TestDeleteBook_Returns204WithEmptyBody(t *testing.T) {
	r := newRouter()
	createBook(t, r, "to-delete", "a")

	w := do(t, r, http.MethodDelete, "/api/v1/books/1", "")

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", w.Body.String())
	}
}

func TestInternalErrorsDoNotLeakDetails(t *testing.T) {
	// 404 responses must not contain internal details.
	w := do(t, newRouter(), http.MethodGet, "/api/v1/books/999", "")
	if strings.Contains(w.Body.String(), "details") {
		t.Errorf("404 response should not contain details: %s", w.Body.String())
	}
}

func TestCRUDFlow(t *testing.T) {
	r := newRouter()

	// Create
	w := do(t, r, http.MethodPost, "/api/v1/books", validBook)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("create: invalid JSON: %v", err)
	}
	rawID, ok := created["id"].(float64)
	if !ok {
		t.Fatalf("create: response has no numeric \"id\" field: %v", created)
	}
	path := fmt.Sprintf("/api/v1/books/%d", uint(rawID))

	// Get
	w = do(t, r, http.MethodGet, path, "")
	if w.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", w.Code)
	}

	// List
	w = do(t, r, http.MethodGet, "/api/v1/books", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", w.Code)
	}
	var list []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("list: invalid JSON array: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("list: expected 1 book, got %d", len(list))
	}

	// Update
	w = do(t, r, http.MethodPut, path, updatedBook)
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Delete
	w = do(t, r, http.MethodDelete, path, "")
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", w.Code)
	}

	// Get after delete
	w = do(t, r, http.MethodGet, path, "")
	if w.Code != http.StatusNotFound {
		t.Errorf("get after delete: expected 404, got %d", w.Code)
	}
}

// The PUT response must contain the stored book id, not zero.
func TestUpdateBook_ResponseKeepsID(t *testing.T) {
	r := newRouter()

	w := do(t, r, http.MethodPost, "/api/v1/books", validBook)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", w.Code)
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("create: invalid JSON: %v", err)
	}
	id, ok := created["id"].(float64)
	if !ok || id == 0 {
		t.Fatalf("create: no valid id in response: %v", created)
	}

	w = do(t, r, http.MethodPut, fmt.Sprintf("/api/v1/books/%d", uint(id)), updatedBook)
	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d", w.Code)
	}
	var updated map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("update: invalid JSON: %v", err)
	}
	if updated["id"] != created["id"] {
		t.Errorf("update response id = %v, want %v", updated["id"], created["id"])
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func requiredFields() map[string]any {
	return map[string]any{"title": "T", "isbn": "I", "author": "A", "category": "C"}
}

func TestCreateBook_EmptyObjectReturns422(t *testing.T) {
	w := do(t, newRouter(), http.MethodPost, "/api/v1/books", `{}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
	if c := errorCode(t, w); c != "validation_error" {
		t.Errorf("expected validation_error, got %q", c)
	}
}

func TestCreateBook_MissingRequiredFieldReturns422(t *testing.T) {
	r := newRouter()
	for field := range requiredFields() {
		t.Run(field, func(t *testing.T) {
			body := map[string]any{}
			for k, v := range requiredFields() {
				if k != field {
					body[k] = v
				}
			}
			w := do(t, r, http.MethodPost, "/api/v1/books", mustJSON(t, body))
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected 422 without %q, got %d", field, w.Code)
			}
		})
	}
}

func TestUpdateBook_MissingRequiredFieldReturns422(t *testing.T) {
	r := newRouter()
	for field := range requiredFields() {
		t.Run(field, func(t *testing.T) {
			body := map[string]any{}
			for k, v := range requiredFields() {
				if k != field {
					body[k] = v
				}
			}
			w := do(t, r, http.MethodPut, "/api/v1/books/1", mustJSON(t, body))
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected 422 without %q, got %d", field, w.Code)
			}
		})
	}
}

func TestCreateBook_OnlyRequiredFieldsIsEnough(t *testing.T) {
	w := do(t, newRouter(), http.MethodPost, "/api/v1/books", mustJSON(t, requiredFields()))
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestCreateBook_ResponseContainsCreatedBook(t *testing.T) {
	w := do(t, newRouter(), http.MethodPost, "/api/v1/books", validBook)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["title"] != "The Go Programming Language" {
		t.Errorf("unexpected Title: %v", got["title"])
	}
	if id, _ := got["id"].(float64); id != 1 {
		t.Errorf("expected ID 1, got %v", got["id"])
	}
}