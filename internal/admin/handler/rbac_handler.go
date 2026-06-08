package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/model"
	"goshop/internal/domain/service"
	"goshop/internal/middleware"
	"goshop/pkg/response"
)

type RBACHandler struct {
	svc *service.RBACService
}

func NewRBACHandler(svc *service.RBACService) *RBACHandler {
	return &RBACHandler{svc: svc}
}

func (h *RBACHandler) ListMenus(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	list, err := h.svc.ListMenus(c.Request.Context())
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *RBACHandler) CreateMenu(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	var req service.CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	m, err := h.svc.CreateMenu(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, m)
}

func (h *RBACHandler) UpdateMenu(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.CreateMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	m, err := h.svc.UpdateMenu(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, m)
}

func (h *RBACHandler) DeleteMenu(c *gin.Context) {
	if !middleware.RequirePlatformView(c) {
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteMenu(c.Request.Context(), id); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	var tenantID *uint64
	if middleware.IsPlatformView(c) {
		scope := c.Query("scope")
		list, err := h.svc.ListRoles(c.Request.Context(), nil, scope)
		if err != nil {
			response.Fail(c, 50000, err.Error())
			return
		}
		response.OK(c, list)
		return
	}
	tid := middleware.GetAdminDataTenantID(c)
	tenantID = &tid
	list, err := h.svc.ListRoles(c.Request.Context(), tenantID, model.RoleScopeTenant)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req service.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if middleware.IsPlatformView(c) {
		if req.Scope != model.RoleScopePlatform {
			req.TenantID = 0
		}
	} else {
		req.Scope = model.RoleScopeTenant
		req.TenantID = middleware.GetAdminDataTenantID(c)
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, role)
}

func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Name   string `json:"name"`
		Status int8   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	role, err := h.svc.UpdateRole(c.Request.Context(), id, req.Name, req.Status)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, role)
}

func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RBACHandler) GetRoleMenus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ids, err := h.svc.GetRoleMenus(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, ids)
}

func (h *RBACHandler) SetRoleMenus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		MenuIDs []uint64 `json:"menu_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if err := h.svc.SetRoleMenus(c.Request.Context(), id, req.MenuIDs); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *RBACHandler) ListAdmins(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var tenantID *uint64
	if middleware.IsPlatformView(c) {
		if tidStr := c.Query("tenant_id"); tidStr != "" {
			tid, _ := strconv.ParseUint(tidStr, 10, 64)
			tenantID = &tid
		}
	} else {
		tid := middleware.GetAdminDataTenantID(c)
		tenantID = &tid
	}
	list, total, err := h.svc.ListAdmins(c.Request.Context(), tenantID, page, pageSize, c.Query("keyword"))
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.Page(c, list, total, page, pageSize)
}

func (h *RBACHandler) CreateAdmin(c *gin.Context) {
	var req service.CreateAdminReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	if middleware.IsPlatformView(c) {
		if req.UserType == "" {
			if req.TenantID == 0 {
				req.UserType = model.AdminUserTypePlatform
			} else {
				req.UserType = model.AdminUserTypeTenant
			}
		}
	} else {
		req.UserType = model.AdminUserTypeTenant
		req.TenantID = middleware.GetAdminDataTenantID(c)
	}
	admin, err := h.svc.CreateAdmin(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, admin)
}

func (h *RBACHandler) UpdateAdmin(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Nickname string   `json:"nickname"`
		Password string   `json:"password"`
		Status   int8     `json:"status"`
		RoleIDs  []uint64 `json:"role_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	admin, err := h.svc.UpdateAdmin(c.Request.Context(), id, req.Nickname, req.Password, req.Status, req.RoleIDs)
	if err != nil {
		response.Fail(c, 40000, err.Error())
		return
	}
	response.OK(c, admin)
}

func (h *RBACHandler) GetAdminRoles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ids, err := h.svc.GetAdminRoleIDs(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, 50000, err.Error())
		return
	}
	response.OK(c, ids)
}
