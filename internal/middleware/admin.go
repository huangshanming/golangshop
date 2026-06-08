package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"goshop/internal/domain/model"
	"goshop/internal/tenant"
	"goshop/pkg/errcode"
	"goshop/pkg/jwtutil"
	"goshop/pkg/response"
)

const (
	UserTypeKey         = "user_type"
	ActiveTenantIDKey   = "active_tenant_id"
	RolesKey            = "roles"
	EffectiveTenantKey  = "effective_tenant_id"
	PlatformViewKey     = "platform_view"
)

func AdminJWT(jwtMgr *jwtutil.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			response.FailWithHTTP(c, http.StatusUnauthorized, errcode.ErrUnauthorized, "缺少 Authorization")
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.FailWithHTTP(c, http.StatusUnauthorized, errcode.ErrUnauthorized, "Authorization 格式错误")
			c.Abort()
			return
		}
		claims, err := jwtMgr.ParseAdminToken(parts[1])
		if err != nil {
			response.FailWithHTTP(c, http.StatusUnauthorized, errcode.ErrTokenInvalid, "Token 无效")
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(TenantIDKey, claims.TenantID)
		c.Set(ActiveTenantIDKey, claims.ActiveTenantID)
		c.Set(UserTypeKey, claims.UserType)
		c.Set(RolesKey, claims.Roles)

		effective := effectiveTenantID(claims)
		c.Set(EffectiveTenantKey, effective)
		c.Set(PlatformViewKey, claims.UserType == model.AdminUserTypePlatform && claims.ActiveTenantID == 0)

		if effective > 0 {
			ctx := tenant.WithTenantID(c.Request.Context(), effective)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

func effectiveTenantID(claims *jwtutil.Claims) uint64 {
	if claims.UserType == model.AdminUserTypePlatform {
		return claims.ActiveTenantID
	}
	return claims.TenantID
}

func GetEffectiveTenantID(c *gin.Context) uint64 {
	if v, ok := c.Get(EffectiveTenantKey); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func IsPlatformView(c *gin.Context) bool {
	if v, ok := c.Get(PlatformViewKey); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func GetUserType(c *gin.Context) string {
	if v, ok := c.Get(UserTypeKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetRoles(c *gin.Context) []string {
	if v, ok := c.Get(RolesKey); ok {
		if roles, ok := v.([]string); ok {
			return roles
		}
	}
	return nil
}

func GetActiveTenantID(c *gin.Context) uint64 {
	if v, ok := c.Get(ActiveTenantIDKey); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func RequireEffectiveTenant(c *gin.Context) bool {
	if GetEffectiveTenantID(c) > 0 {
		return true
	}
	response.Fail(c, errcode.ErrBadRequest, "请先进入租户后再操作")
	return false
}

func RequirePlatformView(c *gin.Context) bool {
	if IsPlatformView(c) {
		return true
	}
	response.FailWithHTTP(c, http.StatusForbidden, errcode.ErrForbidden, "无权限访问")
	return false
}

func CasbinDomain(c *gin.Context) string {
	if IsPlatformView(c) {
		return "0"
	}
	return strconv.FormatUint(GetEffectiveTenantID(c), 10)
}

func HasRole(c *gin.Context, roleCode string) bool {
	for _, r := range GetRoles(c) {
		if r == roleCode {
			return true
		}
	}
	return false
}

func AdminRequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetEffectiveTenantID(c) == 0 {
			response.Fail(c, errcode.ErrBadRequest, "请先进入租户后再操作")
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminRequirePlatformView() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsPlatformView(c) {
			response.FailWithHTTP(c, http.StatusForbidden, errcode.ErrForbidden, "无权限访问")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetAdminDataTenantID 业务数据隔离使用的租户 ID。
func GetAdminDataTenantID(c *gin.Context) uint64 {
	return GetEffectiveTenantID(c)
}
