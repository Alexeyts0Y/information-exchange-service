package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLogger_SetsRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logger())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x?q=1", nil))
	if w.Header().Get(HeaderRequestID) == "" {
		t.Fatal("missing X-Request-ID header")
	}
}

func TestLogger_PreservesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Logger())
	r.GET("/x", func(c *gin.Context) {
		if v, ok := c.Get(ContextRequestID); !ok || v != "fixed" {
			t.Errorf("context request id = %v", v)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(HeaderRequestID, "fixed")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get(HeaderRequestID); got != "fixed" {
		t.Errorf("got %q", got)
	}
}

func TestNewRequestID_Unique(t *testing.T) {
	a := newRequestID()
	b := newRequestID()
	if a == b {
		t.Fatal("ids must be unique")
	}
	if len(a) != 32 {
		t.Errorf("len=%d", len(a))
	}
}
