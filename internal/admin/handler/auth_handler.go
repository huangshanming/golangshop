package handler

import (
	"github.com/gin-gonic/gin"

	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type AuthHandler struct {
	userSvc *service.UserService
}

func NewAuthHandler(userSvc *service.UserService) *AuthHandler {
	return &AuthHandler{userSvc: userSvc}
}

type AdminLoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AdminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	token, admin, err := h.userSvc.AdminLogin(c.Request.Context(), middleware.GetTenantID(c), req.Username, req.Password)
	if err != nil {
		response.Fail(c, 40100, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token, "admin": gin.H{
		"id": admin.ID, "username": admin.Username, "nickname": admin.Nickname, "role": admin.Role,
	}})
}
