package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
	"goshop/pkg/jwtutil"
)

type UserService struct {
	userRepo *repository.UserRepo
	adminRepo *repository.AdminRepo
	jwtMgr   *jwtutil.Manager
}

func NewUserService(userRepo *repository.UserRepo, adminRepo *repository.AdminRepo, jwtMgr *jwtutil.Manager) *UserService {
	return &UserService{userRepo: userRepo, adminRepo: adminRepo, jwtMgr: jwtMgr}
}

func (s *UserService) AppLogin(ctx context.Context, tenantID uint64, phone string) (string, *model.User, error) {
	u, err := s.userRepo.FindByPhone(ctx, tenantID, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		u = &model.User{TenantScoped: model.TenantScoped{TenantID: tenantID}, Phone: phone, Nickname: phone, Status: 1}
		if err := s.userRepo.Create(ctx, u); err != nil {
			return "", nil, err
		}
	} else if err != nil {
		return "", nil, err
	}
	token, err := s.jwtMgr.GenerateAppToken(u.ID, tenantID)
	return token, u, err
}

func (s *UserService) AdminLogin(ctx context.Context, tenantID uint64, username, password string) (string, *model.AdminUser, error) {
	admin, err := s.adminRepo.FindByUsername(ctx, tenantID, username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		admin = &model.AdminUser{
			TenantScoped: model.TenantScoped{TenantID: tenantID},
			Username:     username,
			Password:     string(hash),
			Nickname:     username,
			Role:         "admin",
			Status:       1,
		}
		if err := s.adminRepo.Create(ctx, admin); err != nil {
			return "", nil, err
		}
	} else if err != nil {
		return "", nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	token, err := s.jwtMgr.GenerateAdminToken(admin.ID, tenantID, admin.Username)
	return token, admin, err
}
