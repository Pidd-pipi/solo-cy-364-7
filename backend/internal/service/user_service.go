package service

import (
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/util"
)

// UserService 用户业务逻辑。
type UserService interface {
	Register(username, password, name string, role constants.UserRole, storeID *uint) (*model.User, error)
	Login(username, password string) (*model.User, string, error)
	UpdateProfile(id uint, name string, storeID *uint) (*model.User, error)
	GetByID(id uint) (*model.User, error)
	List(page, pageSize int) ([]model.User, int64, error)
}

type userService struct {
	userRepo  repository.UserRepository
	logger    *slog.Logger
	jwtSecret string
	ttlHours  int
}

// NewUserService 构造用户服务。
func NewUserService(userRepo repository.UserRepository, logger *slog.Logger, jwtSecret string, ttlHours int) UserService {
	return &userService{userRepo: userRepo, logger: logger, jwtSecret: jwtSecret, ttlHours: ttlHours}
}

func (s *userService) Register(username, password, name string, role constants.UserRole, storeID *uint) (*model.User, error) {
	if username == "" || len(password) < 6 {
		return nil, fmt.Errorf("register: %w", util.ErrValidation)
	}
	if !role.Valid() {
		role = constants.RoleStoreManager
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := &model.User{Username: username, PasswordHash: string(hash), Name: name, Role: role, StoreID: storeID}
	if err := s.userRepo.Create(user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, fmt.Errorf("register user[username=%s]: %w", username, util.ErrConflict)
		}
		return nil, fmt.Errorf("register user[username=%s]: %w", username, err)
	}
	s.logger.Info(constants.LogUserRegisterSuccess, "user_id", user.ID, "username", username, "role", role)
	return user, nil
}

func (s *userService) Login(username, password string) (*model.User, string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		s.logger.Warn(constants.LogUserLoginFailed, "username", username, "reason", "user not found")
		return nil, "", fmt.Errorf("login: %w", util.ErrUnauthorized)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		s.logger.Warn(constants.LogUserLoginFailed, "username", username, "reason", "password mismatch")
		return nil, "", fmt.Errorf("login: %w", util.ErrUnauthorized)
	}
	token, err := util.GenerateToken(s.jwtSecret, s.ttlHours, user.ID, user.Username, user.Role, user.StoreID)
	if err != nil {
		return nil, "", fmt.Errorf("login generate token: %w", err)
	}
	s.logger.Info(constants.LogUserLoginSuccess, "user_id", user.ID, "username", username)
	return user, token, nil
}

func (s *userService) UpdateProfile(id uint, name string, storeID *uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("update profile user[id=%d]: %w", id, err)
	}
	if name != "" {
		user.Name = name
	}
	if storeID != nil {
		user.StoreID = storeID
	}
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("update profile user[id=%d]: %w", id, err)
	}
	s.logger.Info(constants.LogUserProfileUpdated, "user_id", id)
	return user, nil
}

func (s *userService) GetByID(id uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user[id=%d]: %w", id, err)
	}
	return user, nil
}

func (s *userService) List(page, pageSize int) ([]model.User, int64, error) {
	users, total, err := s.userRepo.List(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	s.logger.Info(constants.LogUserListQueried, "total", total)
	return users, total, nil
}
