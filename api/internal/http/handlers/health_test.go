package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckReturnsOKWithoutAuth(t *testing.T) {
	handler := NewHealthHandler(nil)
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()

	handler.Check(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q", body)
	}
}
