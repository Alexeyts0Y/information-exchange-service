package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
	"github.com/AlexeytsOY/information-exchange-service/internal/repository"
	"github.com/AlexeytsOY/information-exchange-service/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	MsgInvalidRequestBody  = "invalid request body"
	MsgValidationFailed    = "validation failed"
	MsgInvalidMD5          = "invalid md5"
	MsgInvalidArgument     = "invalid argument"
	MsgThreatNotFound      = "threat not found"
	MsgThreatAlreadyExists = "threat with this md5 already exists"
	MsgVersionConflict     = "version conflict: threat was modified by another request; reload and retry"
	MsgRequestAborted      = "request aborted by client"
	MsgRequestTimedOut     = "request timed out"
	MsgInternalServerError = "internal server error"
	MsgHealthOK            = "ok"
	MsgHealthUnhealthy     = "unhealthy"
)

func ErrorHandler(c *gin.Context, err error, statusCode int) {
	writeError(c, statusCode, err.Error())
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidArgument):
		writeBadRequest(c, MsgInvalidArgument, err.Error())

	case errors.Is(err, repository.ErrNotFound):
		writeError(c, http.StatusNotFound, MsgThreatNotFound)

	case errors.Is(err, repository.ErrAlreadyExists):
		writeError(c, http.StatusConflict, MsgThreatAlreadyExists)

	case errors.Is(err, repository.ErrVersionConflict):
		writeError(c, http.StatusConflict, MsgVersionConflict)

	case errors.Is(err, context.Canceled):
		writeError(c, http.StatusInternalServerError, MsgRequestAborted)

	case errors.Is(err, context.DeadlineExceeded):
		writeError(c, http.StatusInternalServerError, MsgRequestTimedOut)

	default:
		log.Printf("handler: threat: unhandled error: %v", err)
		writeError(c, http.StatusInternalServerError, MsgInternalServerError)
	}
}

func writeError(c *gin.Context, status int, message string) {
	m := message
	c.AbortWithStatusJSON(status, apithreat.ErrorResponse{Message: &m})
}

func writeBadRequest(c *gin.Context, message, details string) {
	m, d := message, details
	c.AbortWithStatusJSON(http.StatusBadRequest, apithreat.ErrorResponse{Message: &m, Details: &d})
}
