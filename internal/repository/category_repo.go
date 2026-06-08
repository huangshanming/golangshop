package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) List(ctx context.Context, tenantID uint64) ([]model.ProductCategory, error) {
	var list []model.ProductCategory
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status = 1", tenantID).
		Order("sort ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *CategoryRepo) GetByID(ctx context.Context, tenantID, id uint64) (*model.ProductCategory, error) {
	var c model.ProductCategory
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&c).Error
	return &c, err
}

func (r *CategoryRepo) Create(ctx context.Context, c *model.ProductCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepo) GetNamesByIDs(ctx context.Context, tenantID uint64, ids []uint64) (map[uint64]string, error) {
	names := make(map[uint64]string)
	if len(ids) == 0 {
		return names, nil
	}
	var list []model.ProductCategory
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND id IN ?", tenantID, ids).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	for _, c := range list {
		names[c.ID] = c.Name
	}
	return names, nil
}
