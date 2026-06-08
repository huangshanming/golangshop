package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type AddressHandler struct {
	svc *service.AddressService
}

func NewAddressHandler(svc *service.AddressService) *AddressHandler {
	return &AddressHandler{svc: svc}
}

func (h *AddressHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *AddressHandler) Create(c *gin.Context) {
	var req service.AddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	a, err := h.svc.Create(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, a)
}

func (h *AddressHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.AddressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), id, req); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *AddressHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), id); err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, nil)
}

type CouponHandler struct {
	svc *service.CouponService
}

func NewCouponHandler(svc *service.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

func (h *CouponHandler) Available(c *gin.Context) {
	goodsAmount, _ := strconv.ParseInt(c.DefaultQuery("goods_amount", "0"), 10, 64)
	list, err := h.svc.ListAvailable(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), goodsAmount)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *CouponHandler) Receive(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Receive(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), id); err != nil {
		response.Fail(c, 40023, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *CouponHandler) Mine(c *gin.Context) {
	list, err := h.svc.ListMine(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

type AfterSaleHandler struct {
	svc *service.AfterSaleService
}

func NewAfterSaleHandler(svc *service.AfterSaleService) *AfterSaleHandler {
	return &AfterSaleHandler{svc: svc}
}

func (h *AfterSaleHandler) Create(c *gin.Context) {
	var req service.CreateAfterSaleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	a, err := h.svc.Create(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, a)
}
