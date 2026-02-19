package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthChecker_SetAndHandler(t *testing.T) {
	h := NewHealthChecker()
	h.Set("whatsapp", "ok")
	h.Set("db", "degraded")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", rec.Header().Get("Content-Type"))
	}
	body := rec.Body.String()
	if body == "" {
		t.Error("response body empty")
	}
	if !strings.Contains(body, "status") {
		t.Error("response should contain status")
	}
	if !strings.Contains(body, "checks") {
		t.Error("response should contain checks")
	}
}

func TestHealthChecker_MethodNotAllowed(t *testing.T) {
	h := NewHealthChecker()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	h.Handler()(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST status = %d, want 405", rec.Code)
	}
}
