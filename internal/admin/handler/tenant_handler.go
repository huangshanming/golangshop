package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type AdminTenantHandler struct {
	svc *service.TenantService
}

func NewAdminTenantHandler(svc *service.TenantService) *AdminTenantHandler {
	return &AdminTenantHandler{svc: svc}
}

func (h *AdminTenantHandler) List(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	list, err := h.svc.ListAdmin(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *AdminTenantHandler) Create(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	var req service.CreateTenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	t, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, t)
}

func (h *AdminTenantHandler) Current(c *gin.Context) {
	tid := middleware.GetAdminDataTenantID(c)
	if tid == 0 {
		response.Fail(c, 40000, "请先进入租户")
		return
	}
	t, err := h.svc.MustGet(c.Request.Context(), tid)
	if err != nil {
		response.Fail(c, 40401, err.Error())
		return
	}
	response.OK(c, t)
}

func (h *AdminTenantHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	t, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 40401, "租户不存在")
		return
	}
	response.OK(c, t)
}
