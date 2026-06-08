package handler

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"goshop/internal/auth"
	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type AuthHandler struct {
	rbacSvc *service.RBACService
	enforcer *casbin.Enforcer
	db       *gorm.DB
}

func NewAuthHandler(rbacSvc *service.RBACService, enforcer *casbin.Enforcer, db *gorm.DB) *AuthHandler {
	return &AuthHandler{rbacSvc: rbacSvc, enforcer: enforcer, db: db}
}

type AdminLoginReq struct {
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password" binding:"required"`
	TenantID *uint64 `json:"tenant_id"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req AdminLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	result, err := h.rbacSvc.AdminLogin(c.Request.Context(), req.Username, req.Password, req.TenantID)
	if err != nil {
		response.Fail(c, 40100, err.Error())
		return
	}
	if result.NeedTenantPick {
		response.OK(c, gin.H{"need_tenant_pick": true, "candidates": result.Candidates})
		return
	}
	h.syncCasbin()
	response.OK(c, gin.H{"token": result.Token, "admin": result.Admin})
}

func (h *AuthHandler) Me(c *gin.Context) {
	me, err := h.rbacSvc.GetMe(
		c.Request.Context(),
		middleware.GetUserID(c),
		middleware.GetUserType(c),
		middleware.GetActiveTenantID(c),
	)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, me)
}

type EnterTenantReq struct {
	TenantID uint64 `json:"tenant_id" binding:"required"`
}

func (h *AuthHandler) EnterTenant(c *gin.Context) {
	var req EnterTenantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	token, admin, err := h.rbacSvc.EnterTenant(c.Request.Context(), middleware.GetUserID(c), req.TenantID)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token, "admin": admin})
}

func (h *AuthHandler) ExitTenant(c *gin.Context) {
	token, admin, err := h.rbacSvc.ExitTenant(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token, "admin": admin})
}

func (h *AuthHandler) syncCasbin() {
	if h.enforcer != nil && h.db != nil {
		_ = auth.SyncPolicies(h.enforcer, h.db)
	}
}
