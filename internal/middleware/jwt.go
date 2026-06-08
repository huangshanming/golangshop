package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"goshop/pkg/errcode"
	"goshop/pkg/jwtutil"
	"goshop/pkg/response"
)

const (
	UserIDKey   = "user_id"
	UsernameKey = "username"
)

func AppJWT(jwtMgr *jwtutil.Manager) gin.HandlerFunc {
	return jwtAuth(jwtMgr.ParseAppToken)
}

func AdminJWT(jwtMgr *jwtutil.Manager) gin.HandlerFunc {
	return jwtAuth(jwtMgr.ParseAdminToken)
}

func jwtAuth(parse func(string) (*jwtutil.Claims, error)) gin.HandlerFunc {
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
		claims, err := parse(parts[1])
		if err != nil {
			response.FailWithHTTP(c, http.StatusUnauthorized, errcode.ErrTokenInvalid, "Token 无效")
			c.Abort()
			return
		}
		headerTenant := GetTenantID(c)
		if headerTenant > 0 && claims.TenantID != headerTenant {
			response.FailWithHTTP(c, http.StatusForbidden, errcode.ErrForbidden, "租户与 Token 不匹配")
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(TenantIDKey, claims.TenantID)
		if claims.Username != "" {
			c.Set(UsernameKey, claims.Username)
		}
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get(UserIDKey); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}
