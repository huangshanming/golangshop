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

type AppLoginReq struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AppLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	token, user, err := h.userSvc.AppLogin(c.Request.Context(), middleware.GetTenantID(c), req.Phone)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token, "user": user})
}
