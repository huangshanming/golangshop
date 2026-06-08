package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type OrderHandler struct {
	orderSvc *service.OrderService
	paySvc   *service.PaymentService
}

func NewOrderHandler(orderSvc *service.OrderService, paySvc *service.PaymentService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc, paySvc: paySvc}
}

func (h *OrderHandler) Preview(c *gin.Context) {
	var req service.PreviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	result, err := h.orderSvc.Preview(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req)
	if err != nil {
		response.Fail(c, mapPreviewError(err), err.Error())
		return
	}
	response.OK(c, result)
}

func mapPreviewError(err error) int {
	msg := err.Error()
	switch {
	case contains(msg, "stock"):
		return 40021
	case contains(msg, "address"):
		return 40022
	case contains(msg, "coupon"):
		return 40023
	default:
		return 40010
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req service.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	order, err := h.orderSvc.Create(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req)
	if err != nil {
		response.Fail(c, mapPreviewError(err), err.Error())
		return
	}
	response.OK(c, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	tab := c.Query("tab")
	list, total, err := h.orderSvc.List(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), page, pageSize, tab)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (h *OrderHandler) Detail(c *gin.Context) {
	orderNo := c.Param("order_no")
	order, err := h.orderSvc.Get(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), orderNo)
	if err != nil {
		response.Fail(c, 40403, "订单不存在")
		return
	}
	response.OK(c, order)
}

func (h *OrderHandler) Pay(c *gin.Context) {
	var req service.PayReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	result, err := h.paySvc.CreatePay(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req)
	if err != nil {
		response.Fail(c, 40020, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *OrderHandler) MockPay(c *gin.Context) {
	orderNo := c.Param("order_no")
	if err := h.paySvc.MockPay(c.Request.Context(), middleware.GetTenantID(c), orderNo); err != nil {
		response.Fail(c, 40020, err.Error())
		return
	}
	response.OK(c, gin.H{"paid": true})
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	orderNo := c.Param("order_no")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.orderSvc.Cancel(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), orderNo, req.Reason); err != nil {
		response.Fail(c, 40010, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *OrderHandler) ConfirmReceive(c *gin.Context) {
	orderNo := c.Param("order_no")
	if err := h.orderSvc.ConfirmReceive(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), orderNo); err != nil {
		response.Fail(c, 40010, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *OrderHandler) Review(c *gin.Context) {
	orderNo := c.Param("order_no")
	if err := h.orderSvc.Review(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), orderNo); err != nil {
		response.Fail(c, 40010, err.Error())
		return
	}
	response.OK(c, nil)
}
