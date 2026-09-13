package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/util"
)

// ErrDuplicate 数据唯一键冲突哨兵错误。
var ErrDuplicate = errors.New("duplicate key")

// UserRepository 用户仓储。
type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	List(page, pageSize int) ([]model.User, int64, error)
	Update(user *model.User) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 构造用户仓储。
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if isDuplicate(err) {
			return fmt.Errorf("create user: %w", ErrDuplicate)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Preload("Store").Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find user by username: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return &u, nil
}

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.Preload("Store").First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find user by id: %w", util.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (r *userRepository) List(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	q := r.db.Model(&model.User{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	if err := q.Preload("Store").Offset((page - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

func (r *userRepository) Update(user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
