package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) Me(c *gin.Context) {
	u, err := h.userSvc.GetProfile(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, 40401, "用户不存在")
		return
	}
	response.OK(c, u)
}

type updateProfileReq struct {
	Nickname string `json:"nickname" binding:"required"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	u, err := h.userSvc.UpdateProfile(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), req.Nickname)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, u)
}

func (h *UserHandler) BalanceLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.userSvc.ListBalanceLogs(c.Request.Context(), middleware.GetTenantID(c), middleware.GetUserID(c), page, pageSize)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}
