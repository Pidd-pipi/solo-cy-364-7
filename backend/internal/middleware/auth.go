package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/config"
	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/util"
)

const CtxUserKey = "user"

// AuthRequired 验证 JWT，将用户信息注入 gin.Context（c.Set("user", ...)）。
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			util.Fail(c, 401, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			util.Fail(c, 401, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		claims, err := util.ParseToken(cfg.JWTSecret, parts[1])
		if err != nil {
			util.GetLogger().Warn(constants.LogAuthRequired, "error", err)
			util.Fail(c, 401, constants.CodeUnauthorized, constants.MsgUnauthorized)
			c.Abort()
			return
		}
		c.Set(CtxUserKey, claims)
		c.Next()
	}
}

// CurrentUser 从 gin.Context 中取当前用户。
func CurrentUser(c *gin.Context) (*util.JWTClaims, error) {
	v, exists := c.Get(CtxUserKey)
	if !exists {
		return nil, errors.New("user not in context")
	}
	claims, ok := v.(*util.JWTClaims)
	if !ok {
		return nil, errors.New("invalid user claims type")
	}
	return claims, nil
}
