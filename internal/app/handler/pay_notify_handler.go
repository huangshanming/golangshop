package handler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type PayNotifyHandler struct {
	paySvc *service.PaymentService
}

func NewPayNotifyHandler(paySvc *service.PaymentService) *PayNotifyHandler {
	return &PayNotifyHandler{paySvc: paySvc}
}

func (h *PayNotifyHandler) Wechat(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		tenantID = 1
	}
	signature := c.GetHeader("Wechatpay-Signature")
	if err := h.paySvc.HandleWechatNotify(c.Request.Context(), tenantID, body, signature); err != nil {
		c.String(500, "FAIL")
		return
	}
	c.String(200, "SUCCESS")
}

func (h *PayNotifyHandler) Alipay(c *gin.Context) {
	values := make(map[string]string)
	if err := c.Request.ParseForm(); err == nil {
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				values[k] = v[0]
			}
		}
	}
	tenantID, _ := strconv.ParseUint(values["tenant_id"], 10, 64)
	if tenantID == 0 {
		tenantID = 1
	}
	if err := h.paySvc.HandleAlipayNotify(c.Request.Context(), tenantID, values); err != nil {
		c.String(500, "fail")
		return
	}
	c.String(200, "success")
}

func (h *PayNotifyHandler) MockNotify(c *gin.Context) {
	var req struct {
		OrderNo string `json:"order_no"`
	}
	_ = c.ShouldBindJSON(&req)
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		tenantID = 1
	}
	if err := h.paySvc.MockPay(c.Request.Context(), tenantID, req.OrderNo); err != nil {
		response.Fail(c, 40020, err.Error())
		return
	}
	response.OK(c, nil)
}
