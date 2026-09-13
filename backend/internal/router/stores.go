package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/handler"
	"github.com/ld/storeinventory/internal/middleware"
)

// registerStoreRoutes 门店路由：建档/修改/删除仅总部与管理员。
func registerStoreRoutes(v1 *gin.RouterGroup, h *handler.StoreHandler, auth gin.HandlerFunc, adminRoles []constants.UserRole) {
	stores := v1.Group("/stores", auth)
	stores.GET("", h.List)
	stores.GET("/all", h.ListAll)
	stores.GET("/:id", h.Get)
	stores.POST("", middleware.RequireRole(adminRoles...), h.Create)
	stores.PUT("/:id", middleware.RequireRole(adminRoles...), h.Update)
	stores.DELETE("/:id", middleware.RequireRole(adminRoles...), h.Delete)
}
