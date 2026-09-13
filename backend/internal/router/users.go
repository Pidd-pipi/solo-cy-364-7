package router

import (
	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/handler"
	"github.com/ld/storeinventory/internal/middleware"
)

// registerUserRoutes 用户路由。
func registerUserRoutes(v1 *gin.RouterGroup, h *handler.UserHandler, auth gin.HandlerFunc) {
	users := v1.Group("/users", auth)
	users.GET("/me", h.Me)
	users.PUT("/me", h.UpdateProfile)
	users.GET("/:id", middleware.RequireRole(constants.RoleAdmin, constants.RoleHQ), h.Get)
	users.GET("", middleware.RequireRole(constants.RoleAdmin, constants.RoleHQ), h.List)
}
