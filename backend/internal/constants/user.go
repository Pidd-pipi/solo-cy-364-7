package constants

import "errors"

// UserRole 用户角色枚举，前后端共享定义（frontend/src/constants/user.ts 对应实现）。
type UserRole string

const (
	RoleHQ           UserRole = "hq"
	RoleStoreManager UserRole = "store_manager"
	RoleAdmin        UserRole = "admin"
)

// ErrInvalidUserRole 非法角色哨兵错误。
var ErrInvalidUserRole = errors.New("invalid user role")

func (r UserRole) Valid() bool {
	switch r {
	case RoleHQ, RoleStoreManager, RoleAdmin:
		return true
	}
	return false
}
