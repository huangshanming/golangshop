package handler

import (
	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/pkg/response"
)

type TenantHandler struct {
	svc *service.TenantService
}

func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

func (h *TenantHandler) ListPublic(c *gin.Context) {
	list, err := h.svc.ListPublic(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}
