package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
	"goshop/internal/repository"
)

type TenantService struct {
	repo *repository.TenantRepo
}

func NewTenantService(repo *repository.TenantRepo) *TenantService {
	return &TenantService{repo: repo}
}

type CreateTenantReq struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

var tenantCodeRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,31}$`)

func (s *TenantService) ListPublic(ctx context.Context) ([]model.Tenant, error) {
	return s.repo.ListActive(ctx)
}

func (s *TenantService) ListAdmin(ctx context.Context) ([]model.Tenant, error) {
	return s.repo.ListAll(ctx)
}

func (s *TenantService) Get(ctx context.Context, id uint64) (*model.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TenantService) Create(ctx context.Context, req CreateTenantReq) (*model.Tenant, error) {
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(strings.ToLower(req.Code))
	if name == "" {
		return nil, errors.New("租户名称不能为空")
	}
	if !tenantCodeRe.MatchString(code) {
		return nil, fmt.Errorf("租户编码需为小写字母开头，仅含小写字母、数字、下划线，2~32 位")
	}
	exists, err := s.repo.ExistsCode(ctx, code, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("租户编码 %s 已存在", code)
	}
	t := &model.Tenant{Name: name, Code: code, Status: 1}
	return t, s.repo.Create(ctx, t)
}

func (s *TenantService) MustGet(ctx context.Context, id uint64) (*model.Tenant, error) {
	t, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("租户不存在")
	}
	return t, err
}
