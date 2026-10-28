package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
	"github.com/AlexeytsOY/information-exchange-service/internal/repository"
	"github.com/AlexeytsOY/information-exchange-service/internal/service"
	"github.com/gin-gonic/gin"
)

type mockSvc struct {
	createFn func(ctx context.Context, req *apithreat.ThreatCreateRequest) (*apithreat.Threat, error)
	getFn    func(ctx context.Context, md5 string) (*apithreat.Threat, error)
	updateFn func(ctx context.Context, md5 string, req *apithreat.ThreatUpdateRequest) (*apithreat.Threat, error)
}

func (m *mockSvc) Create(ctx context.Context, r *apithreat.ThreatCreateRequest) (*apithreat.Threat, error) {
	return m.createFn(ctx, r)
}
func (m *mockSvc) Get(ctx context.Context, md5 string) (*apithreat.Threat, error) {
	return m.getFn(ctx, md5)
}
func (m *mockSvc) Update(ctx context.Context, md5 string, r *apithreat.ThreatUpdateRequest) (*apithreat.Threat, error) {
	return m.updateFn(ctx, md5, r)
}

func newRouter(h *ThreatHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/info", h.CreateThreat)
	r.GET("/info/:md5", func(c *gin.Context) { h.GetInfoByMd5(c, c.Param("md5")) })
	r.PUT("/info/:md5", func(c *gin.Context) { h.UpdateThreat(c, c.Param("md5")) })
	return r
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeErr(t *testing.T, w *httptest.ResponseRecorder) apithreat.ErrorResponse {
	t.Helper()
	var e apithreat.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatalf("decode: %v body=%s", err, w.Body.String())
	}
	return e
}

const validMD5 = "d41d8cd98f00b204e9800998ecf8427e"

func TestNewThreatHandler_PanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewThreatHandler(nil)
}

func TestCreateThreat(t *testing.T) {
	t.Run("201 success", func(t *testing.T) {
		svc := &mockSvc{createFn: func(_ context.Context, r *apithreat.ThreatCreateRequest) (*apithreat.Threat, error) {
			return &apithreat.Threat{Md5: r.Md5, Name: r.Name}, nil
		}}
		w := doJSON(t, newRouter(NewThreatHandler(svc)), http.MethodPost, "/info",
			map[string]any{"md5": validMD5, "name": "n"})
		if w.Code != http.StatusCreated {
			t.Fatalf("status=%d body=%s", w.Code, w.Body)
		}
	})

	t.Run("400 invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/info", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newRouter(NewThreatHandler(&mockSvc{})).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
		if *decodeErr(t, w).Message != MsgInvalidRequestBody {
			t.Errorf("wrong message: %s", w.Body)
		}
	})

	t.Run("400 invalid md5", func(t *testing.T) {
		w := doJSON(t, newRouter(NewThreatHandler(&mockSvc{})), http.MethodPost, "/info",
			map[string]any{"md5": "short", "name": "n"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
		if *decodeErr(t, w).Message != MsgValidationFailed {
			t.Errorf("wrong message: %s", w.Body)
		}
	})

	t.Run("409 already exists", func(t *testing.T) {
		svc := &mockSvc{createFn: func(context.Context, *apithreat.ThreatCreateRequest) (*apithreat.Threat, error) {
			return nil, repository.ErrAlreadyExists
		}}
		w := doJSON(t, newRouter(NewThreatHandler(svc)), http.MethodPost, "/info",
			map[string]any{"md5": validMD5, "name": "n"})
		if w.Code != http.StatusConflict {
			t.Fatalf("status=%d", w.Code)
		}
	})
}

func TestGetInfoByMd5(t *testing.T) {
	t.Run("200 success", func(t *testing.T) {
		svc := &mockSvc{getFn: func(_ context.Context, md5 string) (*apithreat.Threat, error) {
			return &apithreat.Threat{Md5: md5, Name: "n"}, nil
		}}
		w := doJSON(t, newRouter(NewThreatHandler(svc)), http.MethodGet, "/info/"+validMD5, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("400 invalid md5", func(t *testing.T) {
		w := doJSON(t, newRouter(NewThreatHandler(&mockSvc{})), http.MethodGet, "/info/zzz", nil)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		svc := &mockSvc{getFn: func(context.Context, string) (*apithreat.Threat, error) {
			return nil, repository.ErrNotFound
		}}
		w := doJSON(t, newRouter(NewThreatHandler(svc)), http.MethodGet, "/info/"+validMD5, nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status=%d", w.Code)
		}
	})
}

func TestUpdateThreat(t *testing.T) {
	t.Run("200 success", func(t *testing.T) {
		svc := &mockSvc{updateFn: func(_ context.Context, md5 string, r *apithreat.ThreatUpdateRequest) (*apithreat.Threat, error) {
			return &apithreat.Threat{Md5: md5, Name: r.Name}, nil
		}}
		w := doJSON(t, newRouter(NewThreatHandler(svc)), http.MethodPut, "/info/"+validMD5,
			map[string]any{"name": "new"})
		if w.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", w.Code, w.Body)
		}
	})

	t.Run("400 invalid md5", func(t *testing.T) {
		w := doJSON(t, newRouter(NewThreatHandler(&mockSvc{})), http.MethodPut, "/info/xx",
			map[string]any{"name": "new"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("400 invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/info/"+validMD5, bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newRouter(NewThreatHandler(&mockSvc{})).ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
	})

	t.Run("400 empty name", func(t *testing.T) {
		w := doJSON(t, newRouter(NewThreatHandler(&mockSvc{})), http.MethodPut, "/info/"+validMD5,
			map[string]any{"name": ""})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", w.Code, w.Body)
		}
	})

	t.Run("409 version conflict", func(t *testing.T) {
		svc := &mockSvc{updateFn: func(context.Context, string, *apithreat.ThreatUpdateRequest) (*apithreat.Threat, error) {
			return nil, repository.ErrVersionConflict
		}}
		w := doJSON(t, newRouter(NewThreatHandler(svc)), http.MethodPut, "/info/"+validMD5,
			map[string]any{"name": "n"})
		if w.Code != http.StatusConflict {
			t.Fatalf("status=%d", w.Code)
		}
	})
}

func TestWriteServiceError_Unknown(t *testing.T) {
	svc := &mockSvc{getFn: func(context.Context, string) (*apithreat.Threat, error) {
		return nil, errors.New("boom")
	}}
	w := doJSON(t, newRouter(NewThreatHandler(svc)), http.MethodGet, "/info/"+validMD5, nil)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", w.Code)
	}
	if *decodeErr(t, w).Message != MsgInternalServerError {
		t.Errorf("wrong msg: %s", w.Body)
	}
}

var _ = service.ErrInvalidArgument // ensure import used in tests when adding more cases
