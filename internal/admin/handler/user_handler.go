package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type AdminUserHandler struct {
	balanceSvc *service.BalanceService
}

func NewAdminUserHandler(balanceSvc *service.BalanceService) *AdminUserHandler {
	return &AdminUserHandler{balanceSvc: balanceSvc}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.balanceSvc.ListUsers(c.Request.Context(), middleware.GetAdminDataTenantID(c), page, pageSize, c.Query("keyword"))
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

type adjustBalanceReq struct {
	ChangeAmount int64  `json:"change_amount" binding:"required"`
	Remark       string `json:"remark"`
}

func (h *AdminUserHandler) AdjustBalance(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req adjustBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.balanceSvc.Adjust(c.Request.Context(), middleware.GetAdminDataTenantID(c), id, req.ChangeAmount, req.Remark, middleware.GetUserID(c)); err != nil {
		response.Fail(c, 40025, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *AdminUserHandler) BalanceLogs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 64)
	list, total, err := h.balanceSvc.ListLogs(c.Request.Context(), middleware.GetAdminDataTenantID(c), id, int(page), int(pageSize))
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, int(page), int(pageSize))
}

type AdminOrderHandler struct {
	orderSvc *service.OrderService
}

func NewAdminOrderHandler(orderSvc *service.OrderService) *AdminOrderHandler {
	return &AdminOrderHandler{orderSvc: orderSvc}
}

func (h *AdminOrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		v8 := int8(v)
		status = &v8
	}
	list, total, err := h.orderSvc.ListAdmin(c.Request.Context(), middleware.GetAdminDataTenantID(c), page, pageSize, status)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (h *AdminOrderHandler) Ship(c *gin.Context) {
	orderNo := c.Param("order_no")
	var req service.ShipOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.orderSvc.Ship(c.Request.Context(), middleware.GetAdminDataTenantID(c), orderNo, req); err != nil {
		response.Fail(c, 40010, err.Error())
		return
	}
	response.OK(c, nil)
}
