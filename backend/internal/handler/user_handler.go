package handler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/dto"
	"github.com/ld/storeinventory/internal/middleware"
	"github.com/ld/storeinventory/internal/service"
	"github.com/ld/storeinventory/internal/util"
)

// UserHandler 用户接口处理器。
type UserHandler struct {
	userSvc service.UserService
}

// NewUserHandler 构造用户处理器。
func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// Register 注册。
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	user, err := h.userSvc.Register(req.Username, req.Password, req.Name, req.Role, req.StoreID)
	if err != nil {
		c.Error(fmt.Errorf("handler register: %w", err))
		return
	}
	util.OKMessage(c, constants.MsgRegisterSuccess, user)
}

// Login 登录，返回 JWT。
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	user, token, err := h.userSvc.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, util.ErrUnauthorized) {
			c.Error(util.Unauthorized(constants.MsgPasswordIncorrect, err))
			return
		}
		c.Error(fmt.Errorf("handler login: %w", err))
		return
	}
	util.OKMessage(c, constants.MsgLoginSuccess, dto.LoginResponse{Token: token, User: user})
}

// Me 当前用户信息。
func (h *UserHandler) Me(c *gin.Context) {
	claims, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, err))
		return
	}
	user, err := h.userSvc.GetByID(claims.UserID)
	if err != nil {
		c.Error(fmt.Errorf("handler me: %w", err))
		return
	}
	util.OK(c, user)
}

// UpdateProfile 修改资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	claims, err := middleware.CurrentUser(c)
	if err != nil {
		c.Error(util.Unauthorized(constants.MsgUnauthorized, err))
		return
	}
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	user, err := h.userSvc.UpdateProfile(claims.UserID, req.Name, req.StoreID)
	if err != nil {
		c.Error(fmt.Errorf("handler update profile: %w", err))
		return
	}
	util.OK(c, user)
}

// List 用户列表（总部/管理员）。
func (h *UserHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(util.Validation(constants.MsgInvalidRequest, err))
		return
	}
	q.Normalize()
	users, total, err := h.userSvc.List(q.Page, q.PageSize)
	if err != nil {
		c.Error(fmt.Errorf("handler list users: %w", err))
		return
	}
	util.OK(c, gin.H{"list": users, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// Get 用户详情。
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(util.Validation("无效的用户ID", err))
		return
	}
	user, err := h.userSvc.GetByID(uint(id))
	if err != nil {
		c.Error(fmt.Errorf("handler get user: %w", err))
		return
	}
	util.OK(c, user)
}
