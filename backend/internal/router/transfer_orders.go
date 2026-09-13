package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/handler"
	"github.com/ld/storeinventory/internal/middleware"
)

// registerTransferRoutes 调拨单路由：创建启用限流，审批仅总部/管理员。
func registerTransferRoutes(v1 *gin.RouterGroup, h *handler.TransferOrderHandler, auth gin.HandlerFunc, managerRoles []constants.UserRole, limiter gin.HandlerFunc) {
	transfers := v1.Group("/transfers", auth)
	transfers.GET("", h.List)
	transfers.POST("", middleware.RequireRole(managerRoles...), limiter, h.Create)
	transfers.PUT("/:id/confirm", middleware.RequireRole(constants.RoleAdmin, constants.RoleHQ), h.Confirm)
	transfers.PUT("/:id/ship", middleware.RequireRole(managerRoles...), h.Ship)
	transfers.PUT("/:id/receive", middleware.RequireRole(managerRoles...), h.Receive)
	transfers.PUT("/:id/cancel", middleware.RequireRole(managerRoles...), h.Cancel)
}
