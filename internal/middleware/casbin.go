package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"

	"goshop/pkg/errcode"
	"goshop/pkg/response"
)

func CasbinAuth(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		sub := c.GetString(UsernameKey)
		if sub == "" {
			sub = "admin"
		}
		obj := c.Request.URL.Path
		act := c.Request.Method
		dom := CasbinDomain(c)
		ok, err := enforcer.Enforce(sub, obj, act, dom)
		if err != nil {
			response.FailWithHTTP(c, http.StatusInternalServerError, errcode.ErrInternal, "权限校验失败")
			c.Abort()
			return
		}
		if !ok {
			response.FailWithHTTP(c, http.StatusForbidden, errcode.ErrForbidden, "无权限访问")
			c.Abort()
			return
		}
		c.Next()
	}
}
