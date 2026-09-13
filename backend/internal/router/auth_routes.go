package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/handler"
)

// registerAuthRoutes 认证相关路由（注册/登录启用严格限流）。
func registerAuthRoutes(v1 *gin.RouterGroup, h *handler.UserHandler, limiter gin.HandlerFunc) {
	auth := v1.Group("/auth")
	auth.POST("/register", limiter, h.Register)
	auth.POST("/login", limiter, h.Login)
}
