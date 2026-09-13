package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/handler"
	"github.com/ld/storeinventory/internal/middleware"
)

// registerStockRecordRoutes 出入库记录路由。
func registerStockRecordRoutes(v1 *gin.RouterGroup, h *handler.StockRecordHandler, auth gin.HandlerFunc, managerRoles []constants.UserRole) {
	records := v1.Group("/records", auth)
	records.GET("", h.List)
	records.GET("/export", h.Export)
	records.POST("", middleware.RequireRole(managerRoles...), h.Create)
	records.POST("/stocktakes", middleware.RequireRole(managerRoles...), h.CreateStocktake)
	records.GET("/stocktakes", h.ListStocktakes)
	analysis := v1.Group("/analysis", auth)
	analysis.GET("/suggestions", h.Suggestions)
}
