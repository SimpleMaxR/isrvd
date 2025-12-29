package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"isrvd/server/config"
	"isrvd/server/helper"
	"isrvd/server/service"
)

// 认证中间件
func Auth() gin.HandlerFunc {
	var session = service.GetAuthService()

	return func(c *gin.Context) {
		token := helper.GetTokenFromRequest(c)
		if token == "" || !session.ValidateToken(token) {
			helper.RespondError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

// 角色检查中间件
func RequireRole(requiredRole config.Role) gin.HandlerFunc {
	var session = service.GetAuthService()

	return func(c *gin.Context) {
		token := helper.GetTokenFromRequest(c)
		if token == "" {
			helper.RespondError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		sess, exists := session.GetSession(token)
		if !exists {
			helper.RespondError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		// Admin 拥有所有权限
		if sess.Role == config.RoleAdmin {
			c.Next()
			return
		}

		if sess.Role != requiredRole {
			helper.RespondError(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
