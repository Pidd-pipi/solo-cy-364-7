package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/handler"
	"github.com/ld/storeinventory/internal/middleware"
)

// registerSKURoutes SKU 路由：维护接口仅总部/管理员，批量导入启用限流。
func registerSKURoutes(v1 *gin.RouterGroup, h *handler.SKUHandler, auth gin.HandlerFunc, adminRoles []constants.UserRole, limiter gin.HandlerFunc) {
	skus := v1.Group("/skus", auth)
	skus.GET("", h.List)
	skus.POST("", middleware.RequireRole(adminRoles...), h.Create)
	skus.POST("/batch-import", middleware.RequireRole(adminRoles...), limiter, h.BatchImport)
	skus.PUT("/:id", middleware.RequireRole(adminRoles...), h.Update)
	skus.DELETE("/:id", middleware.RequireRole(adminRoles...), h.Delete)
}
