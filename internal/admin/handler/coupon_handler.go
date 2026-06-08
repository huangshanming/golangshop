package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type CouponHandler struct {
	svc *service.CouponService
}

func NewCouponHandler(svc *service.CouponService) *CouponHandler {
	return &CouponHandler{svc: svc}
}

func (h *CouponHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	list, total, err := h.svc.ListAdmin(c.Request.Context(), middleware.GetAdminDataTenantID(c), page, pageSize, keyword)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (h *CouponHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	coupon, err := h.svc.GetAdmin(c.Request.Context(), middleware.GetAdminDataTenantID(c), id)
	if err != nil {
		response.Fail(c, 40401, "优惠券不存在")
		return
	}
	response.OK(c, coupon)
}

func (h *CouponHandler) Create(c *gin.Context) {
	var req service.AdminCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	coupon, err := h.svc.CreateAdmin(c.Request.Context(), middleware.GetAdminDataTenantID(c), req)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, coupon)
}

func (h *CouponHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.AdminCouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	coupon, err := h.svc.UpdateAdmin(c.Request.Context(), middleware.GetAdminDataTenantID(c), id, req)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, coupon)
}

type couponStatusReq struct {
	Status int8 `json:"status" binding:"required"`
}

func (h *CouponHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req couponStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.svc.UpdateStatusAdmin(c.Request.Context(), middleware.GetAdminDataTenantID(c), id, req.Status); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, nil)
}
