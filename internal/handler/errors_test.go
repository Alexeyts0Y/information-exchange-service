package handler

import (
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

func newCtx() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func TestWriteServiceError_Table(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		msg    string
	}{
		{"invalid arg", service.ErrInvalidArgument, http.StatusBadRequest, MsgInvalidArgument},
		{"not found", repository.ErrNotFound, http.StatusNotFound, MsgThreatNotFound},
		{"already exists", repository.ErrAlreadyExists, http.StatusConflict, MsgThreatAlreadyExists},
		{"version conflict", repository.ErrVersionConflict, http.StatusConflict, MsgVersionConflict},
		{"ctx canceled", context.Canceled, http.StatusInternalServerError, MsgRequestAborted},
		{"ctx deadline", context.DeadlineExceeded, http.StatusInternalServerError, MsgRequestTimedOut},
		{"unknown", errors.New("boom"), http.StatusInternalServerError, MsgInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := newCtx()
			writeServiceError(c, tc.err)
			if w.Code != tc.status {
				t.Fatalf("status=%d want=%d", w.Code, tc.status)
			}
			var e apithreat.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
				t.Fatalf("json: %v", err)
			}
			if e.Message == nil || *e.Message != tc.msg {
				t.Errorf("msg=%v want=%q", e.Message, tc.msg)
			}
		})
	}
}

func TestWriteBadRequest(t *testing.T) {
	c, w := newCtx()
	writeBadRequest(c, "msg", "details")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", w.Code)
	}
	var e apithreat.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &e)
	if e.Message == nil || *e.Message != "msg" || e.Details == nil || *e.Details != "details" {
		t.Errorf("got: %+v", e)
	}
}

func TestErrorHandler(t *testing.T) {
	c, w := newCtx()
	ErrorHandler(c, errors.New("some error"), http.StatusBadRequest)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", w.Code)
	}
}
