package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
	"goshop/pkg/jwtutil"
)

type UserService struct {
	userRepo    *repository.UserRepo
	adminRepo   *repository.AdminRepo
	balanceRepo *repository.BalanceRepo
	orderRepo   *repository.OrderRepo
	jwtMgr      *jwtutil.Manager
}

func NewUserService(userRepo *repository.UserRepo, adminRepo *repository.AdminRepo, balanceRepo *repository.BalanceRepo, orderRepo *repository.OrderRepo, jwtMgr *jwtutil.Manager) *UserService {
	return &UserService{userRepo: userRepo, adminRepo: adminRepo, balanceRepo: balanceRepo, orderRepo: orderRepo, jwtMgr: jwtMgr}
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

func (s *UserService) GetProfile(ctx context.Context, tenantID, userID uint64) (*model.User, error) {
	return s.userRepo.FindByID(ctx, tenantID, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, tenantID, userID uint64, nickname string) (*model.User, error) {
	if nickname == "" {
		return nil, errors.New("nickname required")
	}
	if err := s.userRepo.UpdateProfile(ctx, tenantID, userID, nickname); err != nil {
		return nil, err
	}
	return s.userRepo.FindByID(ctx, tenantID, userID)
}

func (s *UserService) ListBalanceLogs(ctx context.Context, tenantID, userID uint64, page, pageSize int) ([]BalanceLogVO, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	logs, total, err := s.balanceRepo.ListLogs(ctx, tenantID, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	vos, err := enrichBalanceLogs(ctx, s.orderRepo, tenantID, logs)
	if err != nil {
		return nil, 0, err
	}
	return vos, total, nil
}
