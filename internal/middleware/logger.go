// Package middleware содержит HTTP-middleware приложения.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	HeaderRequestID  = "X-Request-ID"
	ContextRequestID = "request_id"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Set(ContextRequestID, requestID)
		c.Header(HeaderRequestID, requestID)

		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path += "?" + raw
		}

		log.Printf(
			"http: start request_id=%s method=%s path=%q remote_addr=%s",
			requestID, c.Request.Method, path, c.ClientIP(),
		)

		c.Next()

		status := c.Writer.Status()
		size := max(c.Writer.Size(), 0)
		log.Printf(
			"http: end request_id=%s method=%s path=%q status=%d size=%d latency=%s errors=%q",
			requestID, c.Request.Method, path, status, size,
			time.Since(start), c.Errors.String(),
		)
	}
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "ts-" + time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(b)
}
