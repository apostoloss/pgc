package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"my-solution/pkg/health"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rw := httptest.NewRecorder()

	health.Handler(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.Code)
	}

	contentType := rw.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content-type application/json, got %s", contentType)
	}

	// The response is JSON, order of fields is not guaranteed, so decode it
	var got struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(rw.Body).Decode(&got); err != nil {
		t.Fatalf("could not decode health response: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", got.Status)
	}
	if got.Version != "0.1" {
		t.Errorf("expected version '0.1', got '%s'", got.Version)
	}
}
