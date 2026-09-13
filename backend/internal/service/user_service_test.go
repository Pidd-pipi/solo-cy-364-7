package service

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/ld/storeinventory/internal/constants"
	"github.com/ld/storeinventory/internal/model"
	"github.com/ld/storeinventory/internal/repository"
	"github.com/ld/storeinventory/internal/util"
)

// mockUserRepo 内存版用户仓储，用于 service 单测。
type mockUserRepo struct {
	users map[string]*model.User
	seq   uint
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*model.User)}
}

func (m *mockUserRepo) Create(u *model.User) error {
	if _, ok := m.users[u.Username]; ok {
		return repository.ErrDuplicate
	}
	m.seq++
	u.ID = m.seq
	m.users[u.Username] = u
	return nil
}

func (m *mockUserRepo) FindByUsername(username string) (*model.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, util.ErrNotFound
}

func (m *mockUserRepo) FindByID(id uint) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, util.ErrNotFound
}

func (m *mockUserRepo) List(page, pageSize int) ([]model.User, int64, error) {
	var list []model.User
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, int64(len(list)), nil
}

func (m *mockUserRepo) Update(u *model.User) error {
	m.users[u.Username] = u
	return nil
}

func newTestUserService() UserService {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewUserService(newMockUserRepo(), logger, "test-secret", 72)
}

func TestUserServiceRegister(t *testing.T) {
	svc := newTestUserService()
	tests := []struct {
		name     string
		username string
		password string
		role     constants.UserRole
		wantErr  bool
	}{
		{"valid store manager", "alice", "123456", constants.RoleStoreManager, false},
		{"valid hq", "bob", "123456", constants.RoleHQ, false},
		{"empty username", "", "123456", constants.RoleAdmin, true},
		{"short password", "carol", "123", constants.RoleAdmin, true},
		{"duplicate username", "alice", "123456", constants.RoleAdmin, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := svc.Register(tt.username, tt.password, "测试", tt.role, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got user %+v", u)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Username != tt.username {
				t.Errorf("username = %s, want %s", u.Username, tt.username)
			}
			if u.PasswordHash == "" || u.PasswordHash == tt.password {
				t.Errorf("password not hashed")
			}
		})
	}
}

func TestUserServiceLogin(t *testing.T) {
	svc := newTestUserService()
	_, err := svc.Register("lisa", "123456", "丽萨", constants.RoleStoreManager, nil)
	if err != nil {
		t.Fatalf("seed register failed: %v", err)
	}
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"correct password", "lisa", "123456", false},
		{"wrong password", "lisa", "wrong-pass", true},
		{"unknown user", "nobody", "123456", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, token, err := svc.Login(tt.username, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !errors.Is(err, util.ErrUnauthorized) {
					t.Fatalf("expected ErrUnauthorized, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token == "" {
				t.Fatal("expected token")
			}
		})
	}
}

func TestUserServiceUpdateProfile(t *testing.T) {
	svc := newTestUserService()
	u, err := svc.Register("mike", "123456", "迈克", constants.RoleStoreManager, nil)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	updated, err := svc.UpdateProfile(u.ID, "迈克更新", nil)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "迈克更新" {
		t.Errorf("name = %s, want 迈克更新", updated.Name)
	}
}
