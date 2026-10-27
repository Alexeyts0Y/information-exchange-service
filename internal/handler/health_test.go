package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type mockPinger struct{ err error }

func (m *mockPinger) PingContext(context.Context) error { return m.err }

func TestIsHealthy_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", NewHealthHandler(&mockPinger{}, time.Second).IsHealthy)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != MsgHealthOK {
		t.Errorf("body=%v", body)
	}
}

func TestIsHealthy_Unhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", NewHealthHandler(&mockPinger{err: errors.New("db down")}, time.Second).IsHealthy)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestNewHealthHandler_DefaultTimeout(t *testing.T) {
	h := NewHealthHandler(&mockPinger{}, 0)
	if h.timeout != 2*time.Second {
		t.Errorf("timeout=%v", h.timeout)
	}
}
