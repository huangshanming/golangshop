package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type TenantRepo struct {
	db *gorm.DB
}

func NewTenantRepo(db *gorm.DB) *TenantRepo { return &TenantRepo{db: db} }

func (r *TenantRepo) ListActive(ctx context.Context) ([]model.Tenant, error) {
	var list []model.Tenant
	err := r.db.WithContext(ctx).Where("status = 1").Order("id ASC").Find(&list).Error
	return list, err
}

func (r *TenantRepo) ListAll(ctx context.Context) ([]model.Tenant, error) {
	var list []model.Tenant
	err := r.db.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *TenantRepo) GetByID(ctx context.Context, id uint64) (*model.Tenant, error) {
	var t model.Tenant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	return &t, err
}

func (r *TenantRepo) GetByCode(ctx context.Context, code string) (*model.Tenant, error) {
	var t model.Tenant
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&t).Error
	return &t, err
}

func (r *TenantRepo) Create(ctx context.Context, t *model.Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TenantRepo) ExistsCode(ctx context.Context, code string, excludeID uint64) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Tenant{}).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count > 0, err
}
