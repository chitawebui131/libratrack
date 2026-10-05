package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"libratrack/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func serve(r http.Handler, method, target string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })
	return &buf
}

// ---------- RecoveryMiddleware ----------

func TestRecovery_PanicReturns500WithErrorBody(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("boom") })

	w := serve(r, http.MethodGet, "/boom", nil)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Details any    `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v (body: %s)", err, w.Body.String())
	}
	if resp.Error.Code != "internal_server_error" {
		t.Errorf("unexpected code %q", resp.Error.Code)
	}
	if resp.Error.Message != "Internal server error" {
		t.Errorf("unexpected message %q", resp.Error.Message)
	}
}

func TestRecovery_PanicWithErrorValue(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic(errors.New("db down")) })

	w := serve(r, http.MethodGet, "/boom", nil)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !json.Valid(w.Body.Bytes()) {
		t.Errorf("body is not valid JSON: %s", w.Body.String())
	}
}

func TestRecovery_NoPanicPassesThrough(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "fine") })

	w := serve(r, http.MethodGet, "/ok", nil)

	if w.Code != http.StatusOK || w.Body.String() != "fine" {
		t.Errorf("unexpected response: %d %q", w.Code, w.Body.String())
	}
}

func TestRecovery_ServerKeepsServingAfterPanic(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("boom") })
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	if w := serve(r, http.MethodGet, "/boom", nil); w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if w := serve(r, http.MethodGet, "/ok", nil); w.Code != http.StatusOK {
		t.Errorf("expected 200 after a recovered panic, got %d", w.Code)
	}
}

// ---------- LoggingMiddleware ----------

func TestLogging_LogsMethodPathAndStatus(t *testing.T) {
	buf := captureLog(t)
	r := gin.New()
	r.Use(middleware.LoggingMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	serve(r, http.MethodGet, "/ping", nil)

	if !strings.Contains(buf.String(), "GET /ping 200") {
		t.Errorf("log line not found, got: %q", buf.String())
	}
}

func TestLogging_LogsNotFoundStatus(t *testing.T) {
	buf := captureLog(t)
	r := gin.New()
	r.Use(middleware.LoggingMiddleware())

	serve(r, http.MethodGet, "/nope", nil)

	if !strings.Contains(buf.String(), "GET /nope 404") {
		t.Errorf("log line not found, got: %q", buf.String())
	}
}

func TestLogging_DoesNotLogQueryString(t *testing.T) {
	buf := captureLog(t)
	r := gin.New()
	r.Use(middleware.LoggingMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	serve(r, http.MethodGet, "/ping?token=secret", nil)

	if strings.Contains(buf.String(), "secret") {
		t.Errorf("query string leaked into log: %q", buf.String())
	}
}

func TestLogging_LogsRecoveredPanicAs500(t *testing.T) {
	buf := captureLog(t)
	r := gin.New()
	// Logging is the outer middleware and Recovery the inner one, so a panicking request is logged too.
	r.Use(middleware.LoggingMiddleware(), middleware.RecoveryMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("boom") })

	serve(r, http.MethodGet, "/boom", nil)

	if !strings.Contains(buf.String(), "GET /boom 500") {
		t.Errorf("expected 500 log line, got: %q", buf.String())
	}
}

// ---------- CORSMiddleware ----------

func newCORSRouter(called *bool) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORSMiddleware())
	h := func(c *gin.Context) {
		*called = true
		c.Status(http.StatusOK)
	}
	r.GET("/x", h)
	r.OPTIONS("/x", h)
	return r
}

func TestCORS_NoOriginDefaultsToWildcard(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodGet, "/x", nil)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected *, got %q", got)
	}
}

func TestCORS_ReflectsRequestOrigin(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodGet, "/x", map[string]string{"Origin": "http://example.com"})

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Errorf("expected reflected origin, got %q", got)
	}
}

func TestCORS_EnvOverridesOrigin(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "https://app.libratrack.io")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodGet, "/x", map[string]string{"Origin": "http://evil.com"})

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.libratrack.io" {
		t.Errorf("expected configured origin, got %q", got)
	}
}

func TestCORS_SetsMethodsAndHeaders(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodGet, "/x", nil)

	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("unexpected Allow-Methods: %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization" {
		t.Errorf("unexpected Allow-Headers: %q", got)
	}
}

func TestCORS_PreflightReturns204AndSkipsHandler(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodOptions, "/x", map[string]string{
		"Origin":                        "http://example.com",
		"Access-Control-Request-Method": "POST",
	})

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if called {
		t.Error("handler must not run for preflight requests")
	}
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("preflight response is missing Access-Control-Allow-Origin")
	}
}

func TestCORS_PreflightOnUnregisteredRoute(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	r := gin.New()
	r.Use(middleware.CORSMiddleware())
	r.GET("/only-get", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := serve(r, http.MethodOptions, "/only-get", map[string]string{"Origin": "http://example.com"})

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestCORS_NonPreflightCallsHandler(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodGet, "/x", nil)

	if !called || w.Code != http.StatusOK {
		t.Errorf("handler called=%v, status=%d", called, w.Code)
	}
}

// ---------- Middleware chain as in main.go ----------

func TestChain_PanicResponseKeepsCORSHeaders(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	captureLog(t)
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("boom") })

	w := serve(r, http.MethodGet, "/boom", map[string]string{"Origin": "http://example.com"})

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Error("500 response after panic lost its CORS headers; browsers would hide the error from the frontend")
	}
}

// ---------- Behavior added when fixing the middleware ----------

func TestRecovery_DoesNotLeakPanicValueToClient(t *testing.T) {
	captureLog(t)
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("secret-internal-detail") })

	w := serve(r, http.MethodGet, "/boom", nil)

	body := w.Body.String()
	if strings.Contains(body, "secret-internal-detail") {
		t.Errorf("panic value leaked to the client: %s", body)
	}
	if strings.Contains(body, "details") {
		t.Errorf("response must not contain a details field: %s", body)
	}
}

func TestRecovery_LogsPanicValueAndStackTrace(t *testing.T) {
	buf := captureLog(t)
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("boom-for-log") })

	serve(r, http.MethodGet, "/boom", nil)

	out := buf.String()
	if !strings.Contains(out, "panic recovered") || !strings.Contains(out, "boom-for-log") {
		t.Errorf("panic value was not logged: %q", out)
	}
	if !strings.Contains(out, "goroutine") {
		t.Errorf("stack trace was not logged: %q", out)
	}
}

// Recovery is the outer middleware here (the order used in main.go);
// a panicking request must still produce a log line with status 500.
func TestLogging_LogsPanicWhenRecoveryIsOuter(t *testing.T) {
	buf := captureLog(t)
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware(), middleware.LoggingMiddleware())
	r.GET("/boom", func(c *gin.Context) { panic("boom") })

	w := serve(r, http.MethodGet, "/boom", nil)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(buf.String(), "GET /boom 500") {
		t.Errorf("expected a 500 log line for the panicking request, got: %q", buf.String())
	}
}

func TestCORS_SetsVaryOrigin(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodGet, "/x", map[string]string{"Origin": "http://example.com"})

	if got := w.Header().Get("Vary"); !strings.Contains(got, "Origin") {
		t.Errorf("expected Vary: Origin, got %q", got)
	}
}

func TestCORS_SetsMaxAge(t *testing.T) {
	t.Setenv(middleware.ALLOWED_ORIGIN, "")
	var called bool

	w := serve(newCORSRouter(&called), http.MethodOptions, "/x", map[string]string{"Origin": "http://example.com"})

	if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("expected Max-Age 86400, got %q", got)
	}
}