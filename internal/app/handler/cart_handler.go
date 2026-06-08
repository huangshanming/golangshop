package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type CartHandler struct {
	svc *service.CartService
}

func NewCartHandler(svc *service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

type AddCartReq struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	SKUID     uint64 `json:"sku_id" binding:"required"`
	Quantity  int    `json:"quantity"`
}

func (h *CartHandler) Add(c *gin.Context) {
	var req AddCartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	err := h.svc.Add(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req.ProductID, req.SKUID, req.Quantity)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *CartHandler) Remove(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Remove(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), id); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

type patchCartReq struct {
	Quantity *int  `json:"quantity"`
	Checked  *bool `json:"checked"`
}

func (h *CartHandler) Patch(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req patchCartReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), id, req.Quantity, req.Checked); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

type checkAllReq struct {
	Checked bool `json:"checked"`
}

func (h *CartHandler) CheckAll(c *gin.Context) {
	var req checkAllReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.svc.CheckAll(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req.Checked); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}
