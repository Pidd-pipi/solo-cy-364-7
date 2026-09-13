package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/util"
)

// RequireRole 基于用户角色校验权限，只允许指定角色访问。
func RequireRole(roles ...constants.UserRole) gin.HandlerFunc {
	allowed := make(map[constants.UserRole]bool)
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		claims, err := CurrentUser(c)
		if err != nil {
			util.Fail(c, 401, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		if !allowed[claims.Role] {
			util.GetLogger().Warn(constants.LogRBACDenied, "username", claims.Username, "role", claims.Role)
			util.Fail(c, 403, constants.CodeForbidden, constants.MsgForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
