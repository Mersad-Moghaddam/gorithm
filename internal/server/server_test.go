package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("expected 'ok', got %q", body)
	}
}

func TestIndexPage(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if ctype := rec.Header().Get("Content-Type"); ctype != "text/html; charset=utf-8" {
		t.Errorf("expected text/html; charset=utf-8, got %q", ctype)
	}
}
