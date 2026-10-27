package handler

import (
	"net/http"

	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
	"github.com/AlexeytsOY/information-exchange-service/internal/service"
	"github.com/AlexeytsOY/information-exchange-service/internal/validation"
	"github.com/gin-gonic/gin"
)

var _ apithreat.ServerInterface = (*ThreatHandler)(nil)

type ThreatHandler struct {
	svc service.ThreatService
}

func NewThreatHandler(svc service.ThreatService) *ThreatHandler {
	if svc == nil {
		panic("handler: NewThreatHandler: nil service")
	}
	return &ThreatHandler{svc: svc}
}

func (h *ThreatHandler) CreateThreat(c *gin.Context) {
	var req apithreat.ThreatCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, MsgInvalidRequestBody, err.Error())
		return
	}
	if err := validation.CreateRequest(&req); err != nil {
		writeBadRequest(c, MsgValidationFailed, err.Error())
		return
	}

	dto, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto)
}

func (h *ThreatHandler) GetInfoByMd5(c *gin.Context, md5 string) {
	if err := validation.MD5(md5); err != nil {
		writeBadRequest(c, MsgInvalidMD5, err.Error())
		return
	}

	dto, err := h.svc.Get(c.Request.Context(), md5)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto)
}

func (h *ThreatHandler) UpdateThreat(c *gin.Context, md5 string) {
	if err := validation.MD5(md5); err != nil {
		writeBadRequest(c, MsgInvalidMD5, err.Error())
		return
	}

	var req apithreat.ThreatUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, MsgInvalidRequestBody, err.Error())
		return
	}
	if err := validation.UpdateRequest(&req); err != nil {
		writeBadRequest(c, MsgValidationFailed, err.Error())
		return
	}

	dto, err := h.svc.Update(c.Request.Context(), md5, &req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto)
}
