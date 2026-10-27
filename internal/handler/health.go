package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type HealthHandler struct {
	db      Pinger
	timeout time.Duration
}

func NewHealthHandler(db Pinger, timeout time.Duration) *HealthHandler {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &HealthHandler{db: db, timeout: timeout}
}

func (h *HealthHandler) IsHealthy(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		log.Printf("handler: health: db ping failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": MsgHealthUnhealthy})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": MsgHealthOK})
}
