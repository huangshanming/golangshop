package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/tenant"
	"goshop/pkg/errcode"
	"goshop/pkg/response"
)

const TenantIDKey = "tenant_id"

func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDStr := c.GetHeader("X-Tenant-Id")
		if tenantIDStr == "" {
			response.Fail(c, errcode.ErrTenantMissing, "缺少租户标识 X-Tenant-Id")
			c.Abort()
			return
		}
		tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
		if err != nil || tenantID == 0 {
			response.Fail(c, errcode.ErrTenantMissing, "无效的租户标识")
			c.Abort()
			return
		}
		c.Set(TenantIDKey, tenantID)
		ctx := tenant.WithTenantID(c.Request.Context(), tenantID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GetTenantID(c *gin.Context) uint64 {
	if v, ok := c.Get(TenantIDKey); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}
