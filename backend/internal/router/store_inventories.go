package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/handler"
	"github.com/ld/storeinventory/internal/middleware"
)

// registerInventoryRoutes 门店库存路由。
func registerInventoryRoutes(v1 *gin.RouterGroup, h *handler.StoreInventoryHandler, auth gin.HandlerFunc, managerRoles []constants.UserRole) {
	inv := v1.Group("/inventories", auth)
	inv.GET("", h.List)
	inv.GET("/stats", h.Stats)
	inv.GET("/alerts", h.Alerts)
	inv.POST("/ensure", middleware.RequireRole(managerRoles...), h.Ensure)
	inv.PUT("/:id/safety-stock", middleware.RequireRole(managerRoles...), h.SetSafetyStock)
}
