package repository

import (
	"context"

	"gorm.io/gorm"

	"goshop/internal/domain/model"
)

type AddressRepo struct {
	db *gorm.DB
}

func NewAddressRepo(db *gorm.DB) *AddressRepo { return &AddressRepo{db: db} }

func (r *AddressRepo) List(ctx context.Context, tenantID, userID uint64) ([]model.UserAddress, error) {
	var list []model.UserAddress
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("is_default DESC, id DESC").Find(&list).Error
	return list, err
}

func (r *AddressRepo) GetByID(ctx context.Context, tenantID, userID, id uint64) (*model.UserAddress, error) {
	var a model.UserAddress
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, id).First(&a).Error
	return &a, err
}

func (r *AddressRepo) Create(ctx context.Context, a *model.UserAddress) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if a.IsDefault {
			if err := tx.Model(&model.UserAddress{}).
				Where("tenant_id = ? AND user_id = ?", a.TenantID, a.UserID).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(a).Error
	})
}

func (r *AddressRepo) Update(ctx context.Context, a *model.UserAddress) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if a.IsDefault {
			if err := tx.Model(&model.UserAddress{}).
				Where("tenant_id = ? AND user_id = ?", a.TenantID, a.UserID).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Omit("CreatedAt").Save(a).Error
	})
}

func (r *AddressRepo) Delete(ctx context.Context, tenantID, userID, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, id).
		Delete(&model.UserAddress{}).Error
}

func (r *AddressRepo) GetDefault(ctx context.Context, tenantID, userID uint64) (*model.UserAddress, error) {
	var a model.UserAddress
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND is_default = ?", tenantID, userID, true).First(&a).Error
	if err == gorm.ErrRecordNotFound {
		err = r.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).
			Order("id DESC").First(&a).Error
	}
	return &a, err
}
